package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"stock_assistant/backend/ai_service/biz/provider/llm"
	ai "stock_assistant/backend/ai_service/kitex_gen/ai"
	"stock_assistant/backend/stock_service/kitex_gen/stock/stockservice"

	"github.com/cloudwego/kitex/client"
)

// AIServiceImpl implements the last service interface defined in the IDL.
type AIServiceImpl struct {
	llmProvider llm.Provider
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
		log.Printf("falling back to mock provider")
		return &AIServiceImpl{llmProvider: llm.NewMockProvider()}
	}

	return &AIServiceImpl{
		llmProvider: p,
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
