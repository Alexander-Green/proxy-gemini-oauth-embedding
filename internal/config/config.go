package config

import (
	"encoding/json"
	"fmt"
	"os"

	"go.uber.org/zap"
)

type Config struct {
	GoogleAPIKey         string `json:"google_api_key"`
	GoogleEmbeddingModel string `json:"google_embedding_model"`
	EmbeddingDim         int    `json:"embedding_dim"`
	Port                 string `json:"port"`
	LogPayloads          bool   `json:"log_payloads"`
	ShutdownTimeoutSec   int    `json:"shutdown_timeout_sec"`
	ReadTimeoutSec       int    `json:"read_timeout_sec"`
	WriteTimeoutSec      int    `json:"write_timeout_sec"`
}

func Load(path string) (*Config, error) {
	zap.S().Infof("loading config from: %s", path)

	f, err := os.Open(path)
	if err != nil {
		zap.S().Errorf("failed to open config file: %v", err)
		return nil, fmt.Errorf("open config file: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		zap.S().Errorf("failed to decode config: %v", err)
		return nil, fmt.Errorf("decode config: %w", err)
	}

	if cfg.GoogleAPIKey == "" {
		return nil, fmt.Errorf("google_api_key is required")
	}

	// Set defaults
	if cfg.Port == "" {
		cfg.Port = "1234"
	}
	if cfg.GoogleEmbeddingModel == "" {
		cfg.GoogleEmbeddingModel = "gemini-embedding-2"
	}
	if cfg.EmbeddingDim == 0 {
		cfg.EmbeddingDim = 768
	}
	if cfg.ShutdownTimeoutSec == 0 {
		cfg.ShutdownTimeoutSec = 30
	}
	if cfg.ReadTimeoutSec == 0 {
		cfg.ReadTimeoutSec = 30
	}
	if cfg.WriteTimeoutSec == 0 {
		cfg.WriteTimeoutSec = 30
	}

	zap.S().Infof("config loaded successfully: model=%s, dim=%d, port=%s", cfg.GoogleEmbeddingModel, cfg.EmbeddingDim, cfg.Port)
	return &cfg, nil
}
