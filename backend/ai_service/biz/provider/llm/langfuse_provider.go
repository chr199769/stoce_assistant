package llm

import (
	"context"
	"fmt"
	// "os"

	// langfuse "github.com/langfuse/langfuse-go"
)

type LangfuseManager struct {
	// client *langfuse.Client
}

var GlobalLangfuse *LangfuseManager

func InitLangfuse() error {
	/*
		publicKey := os.Getenv("LANGFUSE_PUBLIC_KEY")
		secretKey := os.Getenv("LANGFUSE_SECRET_KEY")
		host := os.Getenv("LANGFUSE_HOST")

		if publicKey == "" || secretKey == "" {
			// Log warning instead of error, so we don't block startup if not configured
			fmt.Println("Warning: Langfuse credentials not set. Tracing disabled.")
			return nil
		}
		if host == "" {
			host = "http://localhost:3000"
		}

		client := langfuse.New(
			langfuse.WithPublicKey(publicKey),
			langfuse.WithSecretKey(secretKey),
			langfuse.WithHost(host),
		)

		GlobalLangfuse = &LangfuseManager{client: client}
	*/
	fmt.Println("Warning: Langfuse disabled due to build issues.")
	GlobalLangfuse = &LangfuseManager{}
	return nil
}

func GetLangfuse() *LangfuseManager {
	return GlobalLangfuse
}

func (m *LangfuseManager) TracePrediction(ctx context.Context, stockCode string, input string, output string, metadata map[string]interface{}) {
	/*
		if m == nil || m.client == nil {
			return
		}

		// Create a trace
		trace, err := m.client.Trace(ctx, langfuse.TraceAttributes{
			Name: "StockPrediction",
			Metadata: map[string]interface{}{
				"stock_code": stockCode,
				"env":        "dev",
			},
			Input: input,
			Output: output,
		})
		if err != nil {
			fmt.Printf("Failed to create trace: %v\n", err)
			return
		}

		// Add generation span (simplified)
		_, err = trace.Generation(ctx, langfuse.GenerationAttributes{
			Name:   "LLM-Predict",
			Input:  input,
			Output: output,
			Model:  "gpt-4-turbo", // Or get from config
			Metadata: metadata,
		})
		if err != nil {
			fmt.Printf("Failed to add generation: %v\n", err)
		}
	*/
}

func (m *LangfuseManager) Flush() {
	/*
		if m != nil && m.client != nil {
			m.client.Flush()
		}
	*/
}
