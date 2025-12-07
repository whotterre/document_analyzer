package dtos

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenRouterRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type OpenRouterResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

type AnalysisResult struct {
	Summary  string         `json:"summary"`
	Type     string         `json:"type"`
	Metadata map[string]any `json:"metadata"`
}
