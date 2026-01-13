package llm

import (
	"context"
	"stock_assistant/backend/ai_service/kitex_gen/ai"
)

type Provider interface {
	Predict(ctx context.Context, stockCode string, days int32, modelName string) (string, float64, string, error)
	RecognizeImage(ctx context.Context, imageData []byte, modelName string) ([]*ai.RecognizedStock, error)
}
