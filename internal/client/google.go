package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const (
	baseURL = "https://generativelanguage.googleapis.com/v1beta/models"
)

type Client struct {
	apiKey      string
	model       string
	dim         int
	logPayloads bool
	httpClient  *http.Client
}

func NewClient(apiKey, model string, dim int, logPayloads bool) *Client {
	return &Client{
		apiKey:      apiKey,
		model:       model,
		dim:         dim,
		logPayloads: logPayloads,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// EmbedText generates embedding for a single text using Google API
func (c *Client) EmbedText(ctx context.Context, text string) ([]float32, int, error) {
	// Add instruction prefix for text-only tasks
	instruction := "Represent the following text for semantic search:\n\n" + text

	reqBody := map[string]interface{}{
		"model": fmt.Sprintf("models/%s", c.model),
		"content": map[string]interface{}{
			"parts": []map[string]string{
				{"text": instruction},
			},
		},
		"output_dimensionality": c.dim,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:embedContent", baseURL, c.model)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if c.logPayloads {
			zap.S().Infof("google API response status: %d | body: %s", resp.StatusCode, string(body))
		} else {
			zap.S().Infof("google API response status: %d", resp.StatusCode)
		}
		return nil, 0, fmt.Errorf("google API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	if c.logPayloads {
		zap.S().Infof("google API response status: %d | body: %s", resp.StatusCode, string(body))
	} else {
		zap.S().Infof("google API response status: %d", resp.StatusCode)
	}

	var result struct {
		Embedding struct {
			Values []float32 `json:"values"`
		} `json:"embedding"`
		UsageMetadata struct {
			PromptTokenCount int `json:"promptTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		zap.S().Errorf("decode response error: %v, body: %s", err, string(body))
		return nil, 0, fmt.Errorf("decode response: %w", err)
	}

	if len(result.Embedding.Values) == 0 {
		zap.S().Errorf("no embeddings in response, body: %s", string(body))
		return nil, 0, fmt.Errorf("no embeddings in response")
	}

	return result.Embedding.Values, result.UsageMetadata.PromptTokenCount, nil
}
