package models

import (
	"time"
)

type Document struct {
	ID            string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Filename      string    `gorm:"not null" json:"filename"`
	ContentType   string    `json:"content_type"`
	FileSize      int64     `json:"file_size"`
	S3Key         string    `gorm:"not null" json:"-"` 
	ExtractedText string    `gorm:"type:text" json:"text,omitempty"`
	CreatedAt     time.Time `json:"created_at"`

	Summary      string         `gorm:"type:text" json:"summary,omitempty"`
	DocumentType string         `json:"document_type,omitempty"`
	Metadata     map[string]any `gorm:"serializer:json" json:"metadata,omitempty"` // Stores dynamic fields like sender, amount, etc.
}

func (Document) TableName() string {
	return "documents"
}