package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type LangfuseManager struct {
	publicKey string
	secretKey string
	host      string
	client    *http.Client
}

type LangfuseConfig struct {
	PublicKey string
	SecretKey string
	Host      string
}

var GlobalLangfuse *LangfuseManager

func InitLangfuse(cfg *LangfuseConfig) error {
	if cfg == nil {
		return fmt.Errorf("langfuse config is nil")
	}

	publicKey := cfg.PublicKey
	secretKey := cfg.SecretKey
	host := cfg.Host

	if publicKey == "" || secretKey == "" {
		// Log warning instead of error, so we don't block startup if not configured
		fmt.Println("Warning: Langfuse credentials not set. Tracing disabled.")
		return nil
	}
	if host == "" {
		host = "http://localhost:3000"
	}

	GlobalLangfuse = &LangfuseManager{
		publicKey: publicKey,
		secretKey: secretKey,
		host:      host,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	return nil
}

func GetLangfuse() *LangfuseManager {
	return GlobalLangfuse
}

func (m *LangfuseManager) sendRequest(ctx context.Context, path string, body interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s%s", m.host, path)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	// Basic Auth
	auth := m.publicKey + ":" + m.secretKey
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", basicAuth)

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("langfuse api error: %d", resp.StatusCode)
	}

	return nil
}

func (m *LangfuseManager) TracePrediction(ctx context.Context, stockCode string, input string, output string, metadata map[string]interface{}) {
	if m == nil {
		return
	}

	// Run in goroutine to avoid blocking
	go func() {
		// Use a detached context or background context to ensure it completes even if request ctx is cancelled
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		traceID := uuid.New().String()
		now := time.Now()

		// 1. Create Trace
		traceBody := map[string]interface{}{
			"id":     traceID,
			"name":   "StockPrediction",
			"input":  input,
			"output": output,
			"metadata": map[string]interface{}{
				"stock_code": stockCode,
				"env":        "dev",
			},
			"timestamp": now.Format(time.RFC3339),
		}

		if err := m.sendRequest(bgCtx, "/api/public/traces", traceBody); err != nil {
			fmt.Printf("Failed to create trace: %v\n", err)
			return // If trace fails, generation will likely fail or be orphaned
		}

		// 2. Create Generation
		generationBody := map[string]interface{}{
			"traceId":   traceID,
			"name":      "LLM-Predict",
			"startTime": now.Format(time.RFC3339),
			"endTime":   time.Now().Format(time.RFC3339), // Approximate
			"model":     "gpt-4-turbo",                   // Or get from config
			"input":     input,
			"output":    output,
			"metadata":  metadata,
		}

		if err := m.sendRequest(bgCtx, "/api/public/generations", generationBody); err != nil {
			fmt.Printf("Failed to create generation: %v\n", err)
		}
	}()
}

func (m *LangfuseManager) Flush() {
	// No-op for simple http client implementation
	// In a more complex implementation, we might wait for the channel to empty
}
