package model

// OpenAI-compatible request format
type OpenAIEmbeddingRequest struct {
	Input any    `json:"input"` // string or []string
	Model string `json:"model"`
}

// OpenAI-compatible response format
type OpenAIEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object   string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index    int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// Google API request format
type GoogleEmbeddingRequest struct {
	Model string `json:"model"`
	Content struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"content"`
	OutputDimensionality *int `json:"output_dimensionality,omitempty"`
}

// Google API response format
type GoogleEmbeddingResponse struct {
	Embeddings []struct {
		Values []float32 `json:"values"`
	} `json:"embeddings"`
}
