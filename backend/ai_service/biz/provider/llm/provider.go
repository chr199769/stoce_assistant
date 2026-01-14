package llm

import (
	"context"
	"stock_assistant/backend/ai_service/kitex_gen/ai"
	"stock_assistant/backend/ai_service/kitex_gen/stock"
)

type Provider interface {
	Predict(ctx context.Context, stockCode string, days int32, modelName string) (string, float64, string, error)
	RecognizeImage(ctx context.Context, imageData []byte, modelName string) ([]*ai.RecognizedStock, error)
	ReviewMarket(ctx context.Context, sectors []*stock.SectorInfo, limitUps []*stock.LimitUpStock, date string) (*ai.MarketReviewResponse, error)
}
