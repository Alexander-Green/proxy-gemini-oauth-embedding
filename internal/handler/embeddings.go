package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"google-embeddings-proxy/internal/client"
	"google-embeddings-proxy/internal/config"
	"google-embeddings-proxy/internal/model"
)

type EmbeddingsHandler struct {
	cfg    *config.Config
	client *client.Client
}

func NewEmbeddingsHandler(cfg *config.Config, client *client.Client) *EmbeddingsHandler {
	return &EmbeddingsHandler{
		cfg:    cfg,
		client: client,
	}
}

func (h *EmbeddingsHandler) HandleEmbeddings(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()

	var req model.OpenAIEmbeddingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		zap.S().Errorf("[%s] ERROR: failed to parse request: %v", requestID, err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Extract texts from input (string or []string)
	texts := h.extractTexts(req.Input)
	if len(texts) == 0 {
		zap.S().Errorf("[%s] ERROR: no texts in input", requestID)
		http.Error(w, "input is empty", http.StatusBadRequest)
		return
	}

	zap.S().Infof("[%s] INFO: processing %d texts, model: %s, dim: %d", requestID, len(texts), req.Model, h.cfg.EmbeddingDim)

	// Generate embeddings for each text
	embeddings := make([][]float32, len(texts))
	totalTokens := 0
	startTime := time.Now()
	for i, text := range texts {
		emb, tokenCount, err := h.client.EmbedText(r.Context(), text)
		if err != nil {
			zap.S().Errorf("[%s] ERROR: failed to embed text %d: %v", requestID, i, err)
			http.Error(w, "embedding generation failed", http.StatusInternalServerError)
			return
		}

		if len(emb) != h.cfg.EmbeddingDim {
			zap.S().Errorf("[%s] ERROR: embedding dimension mismatch: expected %d, got %d", requestID, h.cfg.EmbeddingDim, len(emb))
			http.Error(w, "embedding dimension mismatch", http.StatusInternalServerError)
			return
		}

		embeddings[i] = emb
		totalTokens += tokenCount
	}
	duration := time.Since(startTime)

	// Build OpenAI-compatible response
	response := model.OpenAIEmbeddingResponse{
		Object: "list",
		Model:  req.Model,
		Usage: struct {
			PromptTokens int `json:"prompt_tokens"`
			TotalTokens  int `json:"total_tokens"`
		}{
			PromptTokens: totalTokens,
			TotalTokens:  totalTokens,
		},
	}

	for i, emb := range embeddings {
		response.Data = append(response.Data, struct {
			Object    string    `json:"object"`
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		}{
			Object:    "embedding",
			Embedding: emb,
			Index:     i,
		})
	}

	zap.S().Infof("[%s] INFO: successfully generated %d embeddings in %s", requestID, len(embeddings), duration)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *EmbeddingsHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"model":  h.cfg.GoogleEmbeddingModel,
		"dim":    h.cfg.EmbeddingDim,
	})
}

func (h *EmbeddingsHandler) extractTexts(input any) []string {
	switch v := input.(type) {
	case string:
		return []string{v}
	case []string:
		return v
	case []interface{}:
		texts := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				texts = append(texts, str)
			}
		}
		return texts
	default:
		return nil
	}
}
