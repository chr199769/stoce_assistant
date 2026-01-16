package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"stock_assistant/backend/ai_service/biz/provider/llm"
	ai "stock_assistant/backend/ai_service/kitex_gen/ai"
	"stock_assistant/backend/ai_service/kitex_gen/stock"
	"stock_assistant/backend/ai_service/kitex_gen/stock/stockservice"

	"github.com/cloudwego/kitex/client"
)

// AIServiceImpl implements the last service interface defined in the IDL.
type AIServiceImpl struct {
	llmProvider llm.Provider
	stockClient stockservice.Client
}

func NewAIServiceImpl() *AIServiceImpl {
	cwd, _ := os.Getwd()
	log.Printf("Current working directory: %s", cwd)

	c, err := stockservice.NewClient("stock_service", client.WithHostPorts("localhost:8888"))
	if err != nil {
		log.Printf("failed to init stock client: %v", err)
	}

	// Try to find config file
	configPath := "conf/llm_config.json"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Try absolute path if relative fails
		configPath = filepath.Join(cwd, "conf/llm_config.json")
		log.Printf("Config not found at relative path, trying: %s", configPath)
	}

	// Read config from file
	fileConfig, err := llm.LoadFileConfig(configPath)
	if err != nil {
		log.Printf("failed to read config file from %s: %v, falling back to empty config", configPath, err)
		fileConfig = nil
	} else {
		log.Printf("Successfully loaded config from %s", configPath)
	}

	p, err := llm.NewLangChainProvider(context.Background(), c, fileConfig)
	if err != nil {
		log.Printf("failed to init langchain provider: %v", err)
		// log.Printf("falling back to mock provider")
		// return &AIServiceImpl{llmProvider: llm.NewMockProvider()}
		// Instead of mock, we return a provider that is nil or error prone?
		// Since NewLangChainProvider now returns error if config is missing.
		// We should probably panic here if we strictly don't want mock.
		// Or return a service that has nil provider and check in methods.
		// Let's allow p to be nil but handle it in methods?
		// No, let's make NewLangChainProvider return a valid provider even if config is empty?
		// But I changed it to return error.
		// So let's just log fatal.
		log.Fatalf("Critical: Failed to init LLM provider and mock is disabled: %v", err)
	}

	return &AIServiceImpl{
		llmProvider: p,
		stockClient: c,
	}
}

// GetPrediction implements the AIServiceImpl interface.
func (s *AIServiceImpl) GetPrediction(ctx context.Context, req *ai.GetPredictionRequest) (resp *ai.GetPredictionResponse, err error) {
	log.Printf("Received prediction request: Code=%s, Model=%s", req.Code, req.Model)
	analysis, confidence, newsSummary, err := s.llmProvider.Predict(ctx, req.Code, req.Days, req.Model)
	if err != nil {
		log.Printf("Prediction failed: %v", err)
		return nil, err
	}

	return &ai.GetPredictionResponse{
		Result_: &ai.PredictionResult_{
			Code:         req.Code,
			Confidence:   confidence,
			Analysis:     analysis,
			NewsSummary_: newsSummary,
		},
	}, nil
}

// ImageRecognition implements the AIServiceImpl interface.
func (s *AIServiceImpl) ImageRecognition(ctx context.Context, req *ai.ImageRecognitionRequest) (resp *ai.ImageRecognitionResponse, err error) {
	log.Printf("Received image recognition request: Model=%s, ImageSize=%d", req.Model, len(req.ImageData))

	stocks, err := s.llmProvider.RecognizeImage(ctx, req.ImageData, req.Model)
	if err != nil {
		log.Printf("Image recognition failed: %v", err)
		return nil, err
	}

	return &ai.ImageRecognitionResponse{
		Stocks: stocks,
	}, nil
}

// MarketReview implements the AIServiceImpl interface.
func (s *AIServiceImpl) MarketReview(ctx context.Context, req *ai.MarketReviewRequest) (resp *ai.MarketReviewResponse, err error) {
	log.Printf("Received market review request: Date=%s", req.Date)

	// 1. Fetch Sector Data
	sectorReq := &stock.GetMarketSectorsRequest{
		Type:  "concept",
		Limit: 20,
	}
	sectorResp, err := s.stockClient.GetMarketSectors(ctx, sectorReq)
	if err != nil {
		log.Printf("Failed to get market sectors: %v", err)
		return nil, err
	}

	// 2. Fetch Limit Up Data
	limitUpReq := &stock.GetLimitUpPoolRequest{
		Date: req.Date,
	}
	limitUpResp, err := s.stockClient.GetLimitUpPool(ctx, limitUpReq)
	if err != nil {
		log.Printf("Failed to get limit up pool: %v", err)
		return nil, err
	}

	// 3. Fetch Dragon Tiger List
	dtReq := &stock.GetDragonTigerListRequest{
		Date: req.Date,
	}
	dtResp, err := s.stockClient.GetDragonTigerList(ctx, dtReq)
	if err != nil {
		log.Printf("Failed to get dragon tiger list: %v", err)
		// Don't fail the whole request, just log and pass nil/empty
		dtResp = &stock.GetDragonTigerListResponse{}
	}

	// 4. Call LLM Provider
	review, err := s.llmProvider.ReviewMarket(ctx, sectorResp.Sectors, limitUpResp.Stocks, dtResp.Items, req.Date)
	if err != nil {
		log.Printf("Failed to generate market review: %v", err)
		return nil, err
	}

	return review, nil
}

// AnalyzeMarket implements the AIServiceImpl interface.
func (s *AIServiceImpl) AnalyzeMarket(ctx context.Context, req *ai.MarketAnalysisRequest) (resp *ai.MarketAnalysisResponse, err error) {
	log.Printf("Received market analysis request: Date=%s", req.Date)

	// 1. Fetch Sector Data
	sectorReq := &stock.GetMarketSectorsRequest{
		Type:  "concept",
		Limit: 20,
	}
	sectorResp, err := s.stockClient.GetMarketSectors(ctx, sectorReq)
	if err != nil {
		log.Printf("Failed to get market sectors: %v", err)
		return nil, err
	}

	// 2. Fetch Limit Up Data
	limitUpReq := &stock.GetLimitUpPoolRequest{
		Date: req.Date,
	}
	limitUpResp, err := s.stockClient.GetLimitUpPool(ctx, limitUpReq)
	if err != nil {
		log.Printf("Failed to get limit up pool: %v", err)
		return nil, err
	}

	// 3. Fetch Dragon Tiger List
	dtReq := &stock.GetDragonTigerListRequest{
		Date: req.Date,
	}
	dtResp, err := s.stockClient.GetDragonTigerList(ctx, dtReq)
	if err != nil {
		log.Printf("Failed to get dragon tiger list: %v", err)
		// Don't fail the whole request, just log and pass nil/empty
		dtResp = &stock.GetDragonTigerListResponse{}
	}

	// 4. Call LLM Provider
	log.Printf("Calling AnalyzeMarket with: Sectors=%d, LimitUps=%d, DTItems=%d", len(sectorResp.Sectors), len(limitUpResp.Stocks), len(dtResp.Items))
	analysis, err := s.llmProvider.AnalyzeMarket(ctx, sectorResp.Sectors, limitUpResp.Stocks, dtResp.Items, req.Date)
	if err != nil {
		log.Printf("Failed to generate market analysis: %v", err)
		return nil, err
	}

	return analysis, nil
}
