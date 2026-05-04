package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"google-embeddings-proxy/internal/client"
	"google-embeddings-proxy/internal/config"
	"google-embeddings-proxy/internal/handler"
)

func main() {
	if err := initLogger(); err != nil {
		fmt.Printf("failed to initialize logger: %v\n", err)
		return
	}

	zap.S().Info("starting google-embeddings-proxy")

	cfg, err := config.Load("config.json")
	if err != nil {
		zap.S().Fatalf("failed to load config: %v", err)
	}

	googleClient := client.NewClient(cfg.GoogleAPIKey, cfg.GoogleEmbeddingModel, cfg.EmbeddingDim)
	embeddingsHandler := handler.NewEmbeddingsHandler(cfg, googleClient)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", embeddingsHandler.HandleHealth)
	mux.HandleFunc("/v1/embeddings", embeddingsHandler.HandleEmbeddings)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  time.Duration(cfg.ReadTimeoutSec) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeoutSec) * time.Second,
	}

	zap.S().Infof("starting server on port %s", cfg.Port)
	zap.S().Infof("model: %s, dimension: %d", cfg.GoogleEmbeddingModel, cfg.EmbeddingDim)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.S().Fatalf("server error: %v", err)
		}
	}()

	gracefulShutdown(srv, cfg.ShutdownTimeoutSec)
}

func gracefulShutdown(srv *http.Server, timeoutSec int) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	zap.S().Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		zap.S().Errorf("server shutdown error: %v", err)
	}

	zap.S().Info("server stopped gracefully")
}

func initLogger() error {
	level := zap.InfoLevel

	const zapCongValue = 100

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "severity",
		NameKey:        "log",
		CallerKey:      "caller",
		MessageKey:     "message",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var defaultZapConfig = zap.Config{
		Level:         zap.NewAtomicLevelAt(level),
		Development:   false,
		DisableCaller: true,
		Sampling: &zap.SamplingConfig{
			Initial:    zapCongValue,
			Thereafter: zapCongValue,
		},
		Encoding:         "json",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	log, err := defaultZapConfig.Build()
	if err != nil {
		return err
	}

	zap.ReplaceGlobals(log)
	zap.RedirectStdLog(log)

	return nil
}
