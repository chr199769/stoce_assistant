package llm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"stock_assistant/backend/ai_service/biz/tool"
	ai "stock_assistant/backend/ai_service/kitex_gen/ai"
	"stock_assistant/backend/stock_service/kitex_gen/stock/stockservice"
	"strings"
	"time"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
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

// IsTradingTime checks if the current time is within A-share trading hours (Mon-Fri 9:15-15:00)
// This is a simplified check and does not account for public holidays.
func IsTradingTime() bool {
	now := time.Now()
	weekday := now.Weekday()

	// 1. Check if it's weekend
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}

	// 2. Check time range 09:15 - 15:00
	// Convert current time to minutes from midnight for easier comparison
	currentMinutes := now.Hour()*60 + now.Minute()
	startMinutes := 9*60 + 15 // 09:15
	endMinutes := 15*60 + 0   // 15:00

	return currentMinutes >= startMinutes && currentMinutes <= endMinutes
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
				// cfg = ModelConfig{Provider: ProviderFake}
				return "", 0, "", fmt.Errorf("provider not found and fake provider disabled")
			}
		}
	} else {
		// No file config, fallback
		// cfg = ModelConfig{Provider: ProviderFake}
		return "", 0, "", fmt.Errorf("no config found and fake provider disabled")
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
	marketTool := tool.NewMarketInfoTool()
	analysisTool := tool.NewStockAnalysisTool()
	t := []tools.Tool{stockTool, newsTool, marketTool, analysisTool}

	// Pre-fetch stock data to ensure accuracy and avoid tool calling failures
	stockData, err := stockTool.Call(ctx, stockCode)
	if err != nil {
		log.Printf("Failed to pre-fetch stock data: %v", err)
		stockData = fmt.Sprintf("Error fetching stock data: %v", err)
	}
	newsData, _ := newsTool.Call(ctx, stockCode)
	marketData, _ := marketTool.Call(ctx, stockCode)
	analysisData, _ := analysisTool.Call(ctx, stockCode)

	// 4. Create Agent
	// ZeroShotReactDescription is good for general purpose tool use
	agent := agents.NewOneShotAgent(llm, t, agents.WithMaxIterations(5))
	executor := agents.NewExecutor(agent)

	// 5. Run Chain
	// Determine Trading Status and Context
	isTrading := IsTradingTime()
	tradingStatusStr := "Market Closed"
	predictionFocus := "Prediction for Next Day & 3 Days"
	timeContextInstruction := `
- Current Status: Market Closed (Inter-day / Weekend)
- Focus: Summarize the full-day performance, analyze Dragon & Tiger List data, and provide an outlook for the next trading day and the next 3 days.
- Order Book Relevance: Low (Snapshot data is less relevant after close).
`

	if isTrading {
		tradingStatusStr = "Intraday Trading (9:15-15:00)"
		predictionFocus = "Prediction for Today's Close & 3 Days"
		timeContextInstruction = `
- Current Status: Intraday Trading (Live Market)
- Focus: Analyze real-time Order Book pressure (Total Buy/Sell), WeiBi/WeiCha, and immediate momentum.
- Order Book Relevance: HIGH. Use it to predict the price trend for the rest of TODAY.
`
	}

	input := fmt.Sprintf(`You are an expert A-share Trader AI (Professional Fund Manager level).
Your task is to provide a comprehensive analysis and prediction for the stock %s.
Current Time Context: %s (%s)

Here is the real-time data for the stock:
[Stock Data]
%s

[Advanced Analysis Data]
(Includes Order Book, Chip Distribution, Industry Info, Dragon Tiger List History, Northbound Funds, Stock Heat, Regulatory Notices)
%s

[Recent News]
%s

[Market & Policy Info]
(Includes Dragon & Tiger List status, and relevant news about Musk, Trump, Policy, Industry)
%s

Process (Professional Trader Logic):
1. **Time Context Check**:
%s

2. **Policy & Macro (The "Sky")**:
   - Identify if the stock aligns with current national strategic directions (e.g., "New Quality Productive Forces", "Low-Altitude Economy").
   - Policy-supported sectors enjoy valuation premiums; suppressed ones must be avoided.

3. **Funds & Chips (The "Ground")**:
   - **Northbound Funds**: Check "Northbound Net Inflow". Is "Smart Money" entering?
   - **Dragon Tiger List**: Analyze if institutional or speculative funds are active.
   - **Chip Distribution**: Check "Winner Rate" and "Cost Range". Is the main force accumulating (low cost) or distributing (high cost)?
   - **Order Book**: Analyze intraday pressure (Total Pending Buy/Sell).

4. **Sentiment & Psychology (The "People")**:
   - **Stock Heat (Guba Rank)**: 
     - Rank soaring (e.g., >500 -> Top 20) = High potential for short-term explosion.
     - Consistently Top 10 = Overheated, risk of correction (reverse thinking).
   - Use "Sheep Flock Effect" logic: Enter at divergence, exit at consensus.

5. **Risk Control (The "Shield" - MANDATORY CHECK)**:
   - **Regulatory Notices**: Check [Regulatory Notices]. If there are recent "Inquiry Letters" (问询函) or "Regulatory Letters" (监管函), this is a **VETO** signal (High Risk).
   - **Volatility Rules**: If the stock has risen >100%% in 10 days, warn about "Special Suspension" (停牌核查) risk.

6. **Prediction**:
   - Provide a prediction for the trend (Up, Down, Neutral) for BOTH the %s.
   - Give a confidence score (0-1).

7. **Conclusion**: Summarize the key logic using the "Trader's Perspective".

Output requirements:
- **Language**: The final answer MUST be in Chinese (Simplified Chinese).
- **Tone**: Professional, objective, insightful (like a senior trader).
- **Structure**:
  - Stock Name & Code
  - Current Price & Status
  - Time Context: %s
  - **Core Logic Analysis**:
    - 🏛️ Policy & Macro (天时)
    - 💰 Funds & Chips (地利) - Include Northbound & Order Book
    - 🗣️ Sentiment (人和) - Include Stock Heat
  - **Risk Assessment (风控)**: Regulatory & Volatility Check
  - **Prediction (%s)**: [Trend] - [Reason]
  - Confidence Score
  - Key Driving Factors

Output your final answer as a detailed analysis in Chinese.
`, stockCode, time.Now().Format("2006-01-02 15:04:05"), tradingStatusStr, stockData, analysisData, newsData, marketData, timeContextInstruction, predictionFocus, tradingStatusStr, predictionFocus)

	res, err := chains.Run(ctx, executor, input)
	if err != nil {
		return "", 0, "", err
	}

	// TODO: Parse confidence from text or use a structured output parser in the future
	return res, 0.85, "See analysis for details", nil
}

func (p *LangChainProvider) RecognizeImage(ctx context.Context, imageData []byte, modelName string) ([]*ai.RecognizedStock, error) {
	// 1. Determine ModelConfig
	var cfg ModelConfig
	if p.fileConfig != nil {
		// Find config by model name in all providers
		found := false
		if modelName != "" {
			for provider, c := range p.fileConfig.Models {
				if c.ModelName == modelName {
					cfg = c
					cfg.Provider = ModelProvider(provider)
					found = true
					break
				}
			}
		}

		if !found {
			var ok bool
			cfg, ok = p.fileConfig.Models[string(p.fileConfig.CurrentProvider)]
			if ok {
				cfg.Provider = p.fileConfig.CurrentProvider
			} else {
				// cfg = ModelConfig{Provider: ProviderFake}
				return nil, fmt.Errorf("provider not found and fake provider disabled")
			}
		}
	} else {
		// cfg = ModelConfig{Provider: ProviderFake}
		return nil, fmt.Errorf("no config found and fake provider disabled")
	}

	log.Printf("Using LLM Provider for Image Recognition: %s, Model: %s", cfg.Provider, cfg.ModelName)

	// if cfg.Provider == ProviderFake {
	// 	// Mock implementation
	// 	time.Sleep(1 * time.Second)
	// 	return []*ai.RecognizedStock{
	// 		{Code: "sh600519", Name: "贵州茅台"},
	// 		{Code: "sz000858", Name: "五粮液"},
	// 	}, nil
	// }

	// 2. Create LLM
	llmClient, err := NewModel(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create llm: %w", err)
	}

	// 3. Prepare Image
	// OpenAI expects base64 image URL: data:image/jpeg;base64,{base64_image}
	// We assume jpeg/png compatibility.
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	imageURL := fmt.Sprintf("data:image/jpeg;base64,%s", base64Image)

	// 4. Create Message
	prompt := `Identify any stock market information in this image. 
If you find stock codes and names, list them. 
Focus on A-share stocks (Shanghai/Shenzhen).
If the image contains a list of stocks, extract all of them.
If the image contains a chart for a specific stock, extract that stock.

Return ONLY a JSON array of objects with "code" and "name" fields.
Ensure the response is a valid JSON array enclosed in square brackets [].
Do NOT add any text before or after the JSON array.
Example: [{"code": "sh600519", "name": "Moutai"}, {"code": "sz000001", "name": "Ping An Bank"}]
If no stocks are found, return empty array [].
Do NOT include any markdown formatting (like ` + "```json" + `). Just the raw JSON string.`

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: prompt},
				llms.ImageURLContent{URL: imageURL},
			},
		},
	}

	// 5. Generate Content
	resp, err := llmClient.GenerateContent(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no content generated")
	}

	content := resp.Choices[0].Content
	log.Printf("Raw LLM Response: %s", content)

	// 6. Parse JSON
	// Clean up potential markdown code blocks if the model ignored instructions
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var stocks []*ai.RecognizedStock
	// Use a temporary struct for unmarshalling to handle potential field mismatch
	var tempStocks []struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}

	if err := json.Unmarshal([]byte(content), &tempStocks); err != nil {
		// Try to use regex if JSON parsing fails
		log.Printf("JSON unmarshal failed: %v, trying regex", err)
		// Simple regex to find codes like sh/sz + 6 digits OR just 6 digits
		re := regexp.MustCompile(`((sh|sz)\d{6})|(\d{6})`)
		matches := re.FindAllString(content, -1)

		uniqueCodes := make(map[string]bool)

		for _, match := range matches {
			code := match
			// Fix stock code prefix if missing
			if len(code) == 6 && !strings.HasPrefix(code, "sh") && !strings.HasPrefix(code, "sz") {
				if strings.HasPrefix(code, "6") {
					code = "sh" + code
				} else if strings.HasPrefix(code, "0") || strings.HasPrefix(code, "3") {
					code = "sz" + code
				}
				// Other cases ignored
			}

			// Avoid duplicates and ensure valid format
			if !uniqueCodes[code] && (strings.HasPrefix(code, "sh") || strings.HasPrefix(code, "sz")) {
				uniqueCodes[code] = true
				stocks = append(stocks, &ai.RecognizedStock{
					Code: code,
					Name: "Unknown", // Cannot reliably extract name via regex without structure
				})
			}
		}

		if len(stocks) == 0 {
			// Only return error if regex also failed to find anything
			return nil, fmt.Errorf("failed to parse stock info: %w", err)
		}
	} else {
		for _, s := range tempStocks {
			code := s.Code
			// Fix stock code prefix if missing
			if len(code) == 6 && !strings.HasPrefix(code, "sh") && !strings.HasPrefix(code, "sz") {
				if strings.HasPrefix(code, "6") {
					code = "sh" + code
				} else if strings.HasPrefix(code, "0") || strings.HasPrefix(code, "3") {
					code = "sz" + code
				}
				// Other cases (e.g. 4, 8) might be Beijing stock exchange or others, ignore for now or default
			}

			stocks = append(stocks, &ai.RecognizedStock{
				Code: code,
				Name: s.Name,
			})
		}
	}

	return stocks, nil
}
