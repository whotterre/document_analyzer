package services

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"whotterre/doculyzer/internal/config"
	"whotterre/doculyzer/internal/dtos"
	"whotterre/doculyzer/internal/models"
	"whotterre/doculyzer/internal/repository"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ledongthuc/pdf"
)

type DocumentService struct {
	Repo     *repository.DocumentRepository
	S3Client *s3.Client
	S3Bucket string
}

func NewDocumentService(repo *repository.DocumentRepository, s3Client *s3.Client, bucket string) *DocumentService {
	return &DocumentService{
		Repo:     repo,
		S3Client: s3Client,
		S3Bucket: bucket,
	}
}

func (s *DocumentService) UploadDocument(file *multipart.FileHeader) (*models.Document, error) {
	if file.Size > 5*1024*1024 {
		return nil, fmt.Errorf("file too large (max 5MB)")
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" && ext != ".docx" {
		return nil, fmt.Errorf("only PDF and DOCX files are supported")
	}

	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("documents/%d_%s", time.Now().Unix(), file.Filename)
	_, err = s.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.S3Bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(fileBytes),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	var text string
	var contentType string

	switch ext {
	case ".pdf":
		text, err = extractTextFromPDFBytes(fileBytes)
		contentType = "application/pdf"
	case ".docx":
		text, err = extractTextFromDOCXBytes(fileBytes)
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}

	if err != nil {
		return nil, fmt.Errorf("failed to extract text: %w", err)
	}

	doc := &models.Document{
		Filename:      file.Filename,
		ContentType:   contentType,
		FileSize:      file.Size,
		S3Key:         key,
		ExtractedText: text,
		CreatedAt:     time.Now(),
	}

	if err := s.Repo.Create(doc); err != nil {
		return nil, err
	}

	return doc, nil
}

func (s *DocumentService) AnalyzeDocument(id string) (*models.Document, error) {
	doc, err := s.Repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if doc.ExtractedText == "" {
		return nil, fmt.Errorf("no text to analyze")
	}

	summary, docType, metadata, err := callOpenRouter(doc.ExtractedText)
	if err != nil {
		return nil, err
	}

	doc.Summary = summary
	doc.DocumentType = docType
	doc.Metadata = metadata

	if err := s.Repo.Update(doc); err != nil {
		return nil, err
	}

	return doc, nil
}

func (s *DocumentService) GetDocument(id string) (*models.Document, error) {
	return s.Repo.GetByID(id)
}

func extractTextFromPDFBytes(data []byte) (string, error) {
	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}

	var textBuilder strings.Builder
	totalPage := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		t, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		textBuilder.WriteString(t)
		textBuilder.WriteString("\n")
	}
	return textBuilder.String(), nil
}

func extractTextFromDOCXBytes(data []byte) (string, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}

	var documentXML *zip.File
	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			documentXML = f
			break
		}
	}

	if documentXML == nil {
		return "", fmt.Errorf("invalid docx: missing word/document.xml")
	}

	rc, err := documentXML.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	// Simple XML parsing to extract text from <w:t> tags
	decoder := xml.NewDecoder(rc)
	var textBuilder strings.Builder
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "t" {
				var s string
				if err := decoder.DecodeElement(&s, &se); err == nil {
					textBuilder.WriteString(s)
					textBuilder.WriteString(" ")
				}
			}
			// Add newlines for paragraphs
			if se.Name.Local == "p" {
				textBuilder.WriteString("\n")
			}
		}
	}

	return strings.TrimSpace(textBuilder.String()), nil
}

func callOpenRouter(text string) (string, string, map[string]any, error) {
	apiKey := config.AppConfig.OpenRouterKey
	if apiKey == "" {
		return "", "", nil, fmt.Errorf("OPENROUTER_API_KEY not set")
	}

	prompt := fmt.Sprintf(`Analyze the following document text and return a JSON object with these fields:
1. "summary": A concise summary.
2. "type": The document type (e.g., Invoice, CV, Report).
3. "metadata": A JSON object with extracted key-value pairs (e.g., date, sender, amount).

Text:
%s`, text)

	reqBody := dtos.OpenRouterRequest{
		Model: "openai/gpt-4o-mini",
		Messages: []dtos.Message{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", nil, err
	}
	defer resp.Body.Close()

	var result dtos.OpenRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", nil, err
	}

	if len(result.Choices) == 0 {
		return "", "", nil, fmt.Errorf("no response from LLM")
	}

	content := result.Choices[0].Message.Content
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")

	var analysis dtos.AnalysisResult

	if err := json.Unmarshal([]byte(content), &analysis); err != nil {
		return content, "Unknown", nil, nil
	}

	return analysis.Summary, analysis.Type, analysis.Metadata, nil
}
