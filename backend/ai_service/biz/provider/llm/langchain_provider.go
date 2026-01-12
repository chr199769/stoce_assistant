package llm

import (
	"context"
	"fmt"
	"log"
	"stock_assistant/backend/ai_service/biz/tool"
	"stock_assistant/backend/stock_service/kitex_gen/stock/stockservice"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/tools"
)

type LangChainProvider struct {
	stockClient stockservice.Client
	fileConfig  *FileConfig
}

func NewLangChainProvider(ctx context.Context, stockClient stockservice.Client, fileConfig *FileConfig) (*LangChainProvider, error) {
	return &LangChainProvider{
		stockClient: stockClient,
		fileConfig:  fileConfig,
	}, nil
}

func (p *LangChainProvider) Predict(ctx context.Context, stockCode string, days int32, modelName string) (string, float64, string, error) {
	// 1. Determine ModelConfig
	var cfg ModelConfig
	if p.fileConfig != nil {
		// Find config by model name in all providers
		found := false
		// If modelName is provided, search for it
		if modelName != "" {
			log.Printf("Searching for model: %s", modelName)
			for provider, c := range p.fileConfig.Models {
				log.Printf("Checking provider: %s, model: %s", provider, c.ModelName)
				if c.ModelName == modelName {
					cfg = c
					cfg.Provider = ModelProvider(provider)
					found = true
					break
				}
			}
		}

		// If not found or not provided, use current provider
		if !found {
			if modelName != "" {
				log.Printf("Model %s not found in config, falling back to current provider", modelName)
			}

			var ok bool
			cfg, ok = p.fileConfig.Models[string(p.fileConfig.CurrentProvider)]
			if ok {
				cfg.Provider = p.fileConfig.CurrentProvider
			} else {
				// Fallback to fake
				cfg = ModelConfig{Provider: ProviderFake}
			}
		}
	} else {
		// No file config, fallback
		cfg = ModelConfig{Provider: ProviderFake}
	}

	log.Printf("Using LLM Provider: %s, Model: %s", cfg.Provider, cfg.ModelName)

	// 2. Create LLM
	llm, err := NewModel(ctx, cfg)
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to create llm: %w", err)
	}

	// 3. Create Tools
	stockTool := tool.NewStockPriceTool(p.stockClient)
	newsTool := tool.NewNewsTool()
	t := []tools.Tool{stockTool, newsTool}

	// Pre-fetch stock data to ensure accuracy and avoid tool calling failures
	stockData, err := stockTool.Call(ctx, stockCode)
	if err != nil {
		log.Printf("Failed to pre-fetch stock data: %v", err)
		stockData = fmt.Sprintf("Error fetching stock data: %v", err)
	}
	newsData, _ := newsTool.Call(ctx, stockCode)

	// 4. Create Agent
	// ZeroShotReactDescription is good for general purpose tool use
	agent := agents.NewOneShotAgent(llm, t, agents.WithMaxIterations(5))
	executor := agents.NewExecutor(agent)

	// 5. Run Chain
	input := fmt.Sprintf(`You are a senior financial analyst AI assistant.
Your task is to provide a comprehensive analysis and prediction for the stock %s for the next 1 day and 3 days.

Here is the real-time data for the stock:
[Stock Data]
%s

[Recent News]
%s

Process:
1. Analyze the provided stock data and news.
2. Provide a prediction for the trend (Up, Down, Neutral) for BOTH the next 1 day and the next 3 days.
3. Give a confidence score (0-1) based on the clarity of signals.
4. Summarize the key factors driving your prediction.

Output requirements:
- **Language**: The final answer MUST be in Chinese (Simplified Chinese).
- **Structure**:
  - Stock Name & Code (Use the data from [Stock Data])
  - Current Price & Status
  - News Analysis
  - Prediction (1 Day): [Trend] - [Reason]
  - Prediction (3 Days): [Trend] - [Reason]
  - Confidence Score
  - Key Driving Factors

Output your final answer as a detailed analysis in Chinese.
`, stockCode, stockData, newsData)

	res, err := chains.Run(ctx, executor, input)
	if err != nil {
		return "", 0, "", err
	}

	// TODO: Parse confidence from text or use a structured output parser in the future
	return res, 0.85, "See analysis for details", nil
}
