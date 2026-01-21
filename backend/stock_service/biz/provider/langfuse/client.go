package langfuse

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type LangfuseManager struct {
	publicKey string
	secretKey string
	host      string
	client    *http.Client
}

type LangfuseConfig struct {
	PublicKey string `json:"public_key"`
	SecretKey string `json:"secret_key"`
	Host      string `json:"base_url"`
}

var GlobalLangfuse *LangfuseManager

func InitLangfuse(cfg *LangfuseConfig) error {
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

// Score sends a score for a trace
func (m *LangfuseManager) Score(ctx context.Context, traceID string, name string, value float64, comment string) error {
	if m == nil {
		return nil
	}

	// Run in goroutine to avoid blocking
	go func() {
		// Use a detached context or background context to ensure it completes even if request ctx is cancelled
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		scoreBody := map[string]interface{}{
			"traceId": traceID,
			"name":    name,
			"value":   value,
			"comment": comment,
			// "timestamp": time.Now().Format(time.RFC3339), // Optional
		}

		if err := m.sendRequest(bgCtx, "/api/public/scores", scoreBody); err != nil {
			fmt.Printf("Failed to create score: %v\n", err)
		} else {
			fmt.Printf("Scored trace %s with %f\n", traceID, value)
		}
	}()
	return nil
}
