package langfuse

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type LangfuseManager struct {
	httpClient *http.Client
	host       string
	publicKey  string
	secretKey  string
}

type LangfuseConfig struct {
	PublicKey string `json:"public_key"`
	SecretKey string `json:"secret_key"`
	Host      string `json:"base_url"`
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
		fmt.Println("Warning: Langfuse credentials not set. Tracing disabled.")
		return nil
	}
	if host == "" {
		host = "http://localhost:3000"
	}

	fmt.Printf("[Langfuse] Initialized with host: %s\n", host)

	GlobalLangfuse = &LangfuseManager{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		host:       host,
		publicKey:  publicKey,
		secretKey:  secretKey,
	}
	return nil
}

func GetLangfuse() *LangfuseManager {
	return GlobalLangfuse
}

func (m *LangfuseManager) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	url := fmt.Sprintf("%s%s", m.host, path)
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	// Basic Auth
	auth := m.publicKey + ":" + m.secretKey
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", basicAuth)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Check for non-2xx status codes (except for GET prompt which handles it separately)
	if resp.StatusCode >= 400 && method != "GET" {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		// We can't access jsonBody here directly because it's scoped to the block above
		// So we just print the response body
		fmt.Printf("[Langfuse Error] Status: %d, Body: %s\n", resp.StatusCode, string(bodyBytes))
		return nil, fmt.Errorf("langfuse api error: status=%d body=%s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

func (m *LangfuseManager) TracePrediction(ctx context.Context, stockCode string, input string, output string, metadata map[string]interface{}) string {
	if m == nil {
		return ""
	}

	traceID := uuid.New().String()

	// Sync create trace to get ID
	// Ideally we should start trace at the beginning of request, but here we do it at the end
	// For spans to work, we need traceID.
	return traceID
}

func (m *LangfuseManager) CreateTrace(ctx context.Context, name string, metadata map[string]interface{}) string {
	if m == nil {
		return ""
	}
	traceID := uuid.New().String()
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		now := time.Now().UTC().Format(time.RFC3339)
		traceBody := map[string]interface{}{
			"id":        traceID,
			"name":      name,
			"metadata":  metadata,
			"timestamp": now,
		}
		if resp, err := m.doRequest(bgCtx, "POST", "/api/public/traces", traceBody); err == nil {
			resp.Body.Close()
		} else {
			fmt.Printf("Failed to create trace: %v\n", err)
		}
	}()
	return traceID
}

func (m *LangfuseManager) UpdateTrace(ctx context.Context, traceID string, input interface{}, output interface{}) {
	if m == nil {
		return
	}
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		traceBody := map[string]interface{}{
			"id":     traceID,
			"input":  input,
			"output": output,
		}
		if resp, err := m.doRequest(bgCtx, "POST", "/api/public/traces", traceBody); err == nil {
			resp.Body.Close()
		} else {
			fmt.Printf("Failed to update trace: %v\n", err)
		}
	}()
}

func (m *LangfuseManager) CreateGeneration(ctx context.Context, traceID string, name, model string, input, output interface{}, metadata map[string]interface{}, startTime, endTime time.Time) {
	if m == nil {
		return
	}
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		genID := uuid.New().String()
		genBody := map[string]interface{}{
			"id":        genID,
			"traceId":   traceID,
			"name":      name,
			"startTime": startTime.UTC().Format(time.RFC3339),
			"endTime":   endTime.UTC().Format(time.RFC3339),
			"model":     model,
			"input":     input,
			"output":    output,
			"metadata":  metadata,
		}
		if resp, err := m.doRequest(bgCtx, "POST", "/api/public/generations", genBody); err == nil {
			resp.Body.Close()
		} else {
			fmt.Printf("Failed to create generation: %v\n", err)
		}
	}()
}

// Score adds a score to a trace
func (m *LangfuseManager) Score(ctx context.Context, traceID string, name string, value float64, comment string) error {
	if m == nil {
		return fmt.Errorf("langfuse not initialized")
	}

	scoreBody := map[string]interface{}{
		"id":      uuid.New().String(),
		"traceId": traceID,
		"name":    name,
		"value":   value,
		"comment": comment,
	}

	resp, err := m.doRequest(ctx, "POST", "/api/public/scores", scoreBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to score: status %d", resp.StatusCode)
	}
	return nil
}

// PromptResponse matches Langfuse API response
type PromptResponse struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
	Prompt  string `json:"prompt"`
	Config  struct {
		Model       string  `json:"model"`
		Temperature float64 `json:"temperature"`
	} `json:"config"`
}

// GetPrompt fetches a prompt from Langfuse via REST API
func (m *LangfuseManager) GetPrompt(ctx context.Context, name string, version *int) (string, error) {
	if m == nil {
		return "", fmt.Errorf("langfuse not initialized")
	}

	path := fmt.Sprintf("/api/public/prompts?name=%s", name)
	if version != nil {
		path = fmt.Sprintf("%s&version=%d", path, *version)
	}

	resp, err := m.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get prompt: status %d, body: %s", resp.StatusCode, string(body))
	}

	var promptResp PromptResponse
	if err := json.NewDecoder(resp.Body).Decode(&promptResp); err != nil {
		return "", err
	}

	return promptResp.Prompt, nil
}

// Span creates a span in Langfuse
func (m *LangfuseManager) Span(ctx context.Context, traceID string, parentObservationID *string, name string, input interface{}, output interface{}, startTime, endTime time.Time) error {
	if m == nil {
		return fmt.Errorf("langfuse not initialized")
	}

	spanID := uuid.New().String()
	spanBody := map[string]interface{}{
		"id":        spanID,
		"traceId":   traceID,
		"type":      "SPAN",
		"name":      name,
		"startTime": startTime.UTC().Format(time.RFC3339),
		"endTime":   endTime.UTC().Format(time.RFC3339),
		"input":     input,
		"output":    output,
	}

	if parentObservationID != nil {
		spanBody["parentObservationId"] = *parentObservationID
	}

	// Async execution
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if resp, err := m.doRequest(bgCtx, "POST", "/api/public/spans", spanBody); err == nil {
			resp.Body.Close()
		} else {
			fmt.Printf("Failed to create span %s: %v\n", name, err)
		}
	}()

	return nil
}
