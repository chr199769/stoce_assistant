package llm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"stock_assistant/backend/ai_service/biz/tool"
	ai "stock_assistant/backend/ai_service/kitex_gen/ai"
	"stock_assistant/backend/ai_service/kitex_gen/stock"
	"stock_assistant/backend/ai_service/kitex_gen/stock/stockservice"
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
	marketTool := tool.NewMarketInfoTool()
	analysisTool := tool.NewStockAnalysisTool()
	sectorTool := tool.NewSectorTool(p.stockClient)
	dtTool := tool.NewDragonTigerTool()
	t := []tools.Tool{stockTool, marketTool, analysisTool, sectorTool, dtTool}

	// Pre-fetch stock data to ensure accuracy and avoid tool calling failures
	stockData, err := stockTool.Call(ctx, stockCode)
	if err != nil {
		log.Printf("Failed to pre-fetch stock data: %v", err)
		stockData = fmt.Sprintf("Error fetching stock data: %v", err)
	}

	// Pre-fetch market info (news + dragon tiger + trends) using the unified MarketInfoTool
	marketInfo, _ := marketTool.Call(ctx, stockCode)

	analysisData, _ := analysisTool.Call(ctx, stockCode)

	// Pre-fetch Macro Context
	sectorContext, _ := sectorTool.Call(ctx, "industry")
	dtContext, _ := dtTool.Call(ctx, "") // Get today's general list

	// 4. Create Agent
	// ZeroShotReactDescription is good for general purpose tool use
	agent := agents.NewOneShotAgent(llm, t, agents.WithMaxIterations(5))
	executor := agents.NewExecutor(agent)

	// 5. Run Chain
	// Determine Trading Status and Context
	isTrading := IsTradingTime()
	tradingStatusStr := "已收盘"
	predictionFocus := "次日及未来3日预测"
	timeContextInstruction := `
- Current Status: Market Closed (Inter-day / Weekend)
- Focus: Summarize the full-day performance, analyze Dragon & Tiger List data, and provide an outlook for the next trading day and the next 3 days.
- Order Book Relevance: Low (Snapshot data is less relevant after close).
`

	if isTrading {
		tradingStatusStr = "盘中交易 (9:15-15:00)"
		predictionFocus = "当日收盘及未来3日预测"
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
(Includes Order Book, Chip Distribution, Industry Info, Dragon Tiger List History, Stock Heat, Regulatory Notices)
%s

[Market Intelligence]
(Includes Recent Stock News, Dragon & Tiger Status, Social Trends, and General Market/Policy News)
%s

[Macro & Hot Money Context]
(Includes Top Industries and Today's Dragon Tiger List Overview)
[Top Industries]:
%s

[Today's Hot Money (Dragon Tiger List)]:
%s

Process (Professional Trader Logic):
1. **Time Context Check**:
%s

2. **Policy & Macro (The "Sky")**:
   - Identify if the stock aligns with current national strategic directions.
   - Check if the stock's industry is in the [Top Industries] list. If yes, it's a "Main Line" stock (High Potential).
   - Policy-supported sectors enjoy valuation premiums.

3. **Funds & Chips (The "Ground")**:
   - **Dragon Tiger List**: 
     - Check if the stock is on [Today's Hot Money] list.
     - Check "Dragon Tiger List History" in [Advanced Analysis Data].
     - Identify if "Hot Money" (e.g., Zhao Laoge, Lasa Tiantuan) or "Institutions" are buying.
   - **Chip Distribution**: Check "Winner Rate" and "Cost Range".
   - **Order Book**: Analyze intraday pressure.

4. **Sentiment & Psychology (The "People")**:
   - **Stock Heat (Guba Rank)**: Rank soaring = Short-term explosion.
   - **Sector Resonance**: Does the stock move with its sector leaders?

5. **Risk Control (The "Shield" - MANDATORY CHECK)**:
   - **Regulatory Notices**: Check for Inquiry Letters/Regulatory Letters.
   - **Volatility Rules**: Check for recent abnormal fluctuations.

6. **Prediction**:
   - Provide a prediction for the trend (Up, Down, Neutral) for BOTH the %s.
   - Give a confidence score (0-1).

7. **Conclusion**: Summarize the key logic using the "Trader's Perspective".

Output requirements:
- **Language**: The final answer MUST be in Chinese (Simplified Chinese).
- **Tone**: Professional, objective, insightful.
- **Structure**:
  - 股票名称与代码
  - 当前价格与状态
  - 时间背景: %s
  - **核心逻辑分析**:
    - 🏛️ 政策与宏观 (天时) - 包含板块效应分析
    - 💰 资金与筹码 (地利) - 包含龙虎榜游资分析
    - 🗣️ 情绪与心理 (人和) - 包含个股热度与联动
  - **风控评估 (风控)**: 监管与波动率检查
  - **走势预测 (%s)**: [趋势] - [理由]
  - 置信度评分: [0-1]
  - 关键驱动因素

IMPORTANT: After the detailed analysis, you MUST output a metadata block separated by "---METADATA---".
The metadata block MUST be a valid JSON object with the following fields:
- "confidence": (float) The same confidence score as in the analysis (0.0 to 1.0).
- "news_summary": (string) A concise summary of the most important news/events driving the prediction (max 50 words).

Example Output:
Final Answer:
... (Analysis Text) ...

---METADATA---
{"confidence": 0.85, "news_summary": "Policy support for low-altitude economy and 5G drives positive outlook despite short-term selling pressure."}

Output your final answer starting with "Final Answer:", followed by the detailed analysis in Chinese, and then the metadata block.
`, stockCode, time.Now().Format("2006-01-02 15:04:05"), tradingStatusStr, stockData, analysisData, marketInfo, sectorContext, dtContext, timeContextInstruction, predictionFocus, tradingStatusStr, predictionFocus)

	res, err := chains.Run(ctx, executor, input)
	if err != nil {
		// Fallback: If LangChain fails to parse the output but the model actually returned the content
		// (common with "unable to parse agent output" error), we try to extract it.
		if strings.Contains(err.Error(), "unable to parse agent output") {
			log.Printf("LangChain parse error, attempting to recover content from error message")
			// The error message format is usually: "unable to parse agent output: <actual_output>"
			prefix := "unable to parse agent output: "
			errMsg := err.Error()
			if idx := strings.Index(errMsg, prefix); idx != -1 {
				res = errMsg[idx+len(prefix):]
				// Proceed to parsing
			} else {
				return "", 0, "", err
			}
		} else {
			return "", 0, "", err
		}
	}

	// Parse Output
	analysis := res
	confidence := 0.5 // Default
	newsSummary := "See analysis for details"

	parts := strings.Split(res, "---METADATA---")
	if len(parts) > 1 {
		analysis = strings.TrimSpace(parts[0])
		metadataJSON := strings.TrimSpace(parts[1])
		// Clean JSON (remove potential markdown code blocks)
		metadataJSON = strings.TrimPrefix(metadataJSON, "```json")
		metadataJSON = strings.TrimPrefix(metadataJSON, "```")
		metadataJSON = strings.TrimSuffix(metadataJSON, "```")
		metadataJSON = strings.TrimSpace(metadataJSON)

		var metadata struct {
			Confidence  float64 `json:"confidence"`
			NewsSummary string  `json:"news_summary"`
		}
		if err := json.Unmarshal([]byte(metadataJSON), &metadata); err == nil {
			confidence = metadata.Confidence
			newsSummary = metadata.NewsSummary
		} else {
			log.Printf("Failed to parse metadata JSON: %v. JSON: %s", err, metadataJSON)
		}
	} else {
		log.Printf("Metadata separator not found in response")
	}

	return analysis, confidence, newsSummary, nil
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

	// 2. Create LLM
	llmClient, err := NewModel(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create llm: %w", err)
	}

	// 3. Prepare Image
	// Detect content type
	mimeType := http.DetectContentType(imageData)
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	imageURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Image)

	// 4. Create Message
	prompt := `请识别这张图片中的股票市场信息。
如果发现股票代码和名称，请列出它们。
重点关注A股（上海/深圳）。
如果图片包含股票列表，请提取所有股票。
如果图片是某只股票的走势图，请提取该股票。

仅返回一个包含 "code" 和 "name" 字段的 JSON 对象数组。
确保响应是方括号 [] 括起来的有效 JSON 数组。
不要在 JSON 数组前后添加任何文本。
示例：[{"code": "sh600519", "name": "贵州茅台"}, {"code": "sz000001", "name": "平安银行"}]
如果未找到股票，则返回空数组 []。
不要包含任何 markdown 格式（如 ` + "```json" + `）。只返回原始 JSON 字符串。`

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
			// Return empty list instead of error to avoid 500
			log.Printf("Failed to parse stock info from image (JSON & Regex failed): %v", err)
			return []*ai.RecognizedStock{}, nil
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

func (p *LangChainProvider) ReviewMarket(ctx context.Context, sectors []*stock.SectorInfo, limitUps []*stock.LimitUpStock, dragonTigerList []*stock.DragonTigerItem, date string) (*ai.MarketReviewResponse, error) {
	// ... (Implementation for MarketReview - Focus on Today's Summary)
	// 1. Determine ModelConfig
	var cfg ModelConfig
	if p.fileConfig != nil {
		var ok bool
		cfg, ok = p.fileConfig.Models[string(p.fileConfig.CurrentProvider)]
		if ok {
			cfg.Provider = p.fileConfig.CurrentProvider
		} else {
			return nil, fmt.Errorf("provider not found")
		}
	} else {
		return nil, fmt.Errorf("no config found")
	}

	log.Printf("Using LLM Provider for Market Review: %s, Model: %s", cfg.Provider, cfg.ModelName)

	// 2. Create LLM
	llmClient, err := NewModel(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create llm: %w", err)
	}

	// 3. Prepare Data Context
	var sectorSummary strings.Builder
	sectorSummary.WriteString("Top Sectors:\n")
	for i, s := range sectors {
		if i >= 10 { // Top 10
			break
		}
		sectorSummary.WriteString(fmt.Sprintf("- %s: +%.2f%% (Net Inflow: %.2f), Top Stock: %s\n", s.Name, s.ChangePercent, s.NetInflow, s.TopStockName))
	}

	var limitUpSummary strings.Builder
	limitUpSummary.WriteString(fmt.Sprintf("Limit Up Pool (Total: %d):\n", len(limitUps)))
	// Simple stats
	typeCount := make(map[string]int)
	for _, s := range limitUps {
		typeCount[s.LimitUpType]++
		limitUpSummary.WriteString(fmt.Sprintf("- %s: %s, %s, %s\n", s.Name, s.LimitUpType, s.Reason, s.ChangePercent))
	}

	var dtSummary strings.Builder
	dtSummary.WriteString(fmt.Sprintf("Dragon Tiger List (Top 5 Net Buy):\n"))
	for i, item := range dragonTigerList {
		if i >= 5 {
			break
		}
		dtSummary.WriteString(fmt.Sprintf("- %s: +%.2f%%, Net: %.1f Wan, Reason: %s\n", item.Name, item.ChangePercent, item.NetInflow/10000, item.Reason))
		// Add seats if available (Top 3)
		if len(item.BuySeats) > 0 {
			dtSummary.WriteString("  [Buy Seats]: ")
			for k, seat := range item.BuySeats {
				if k >= 2 {
					break
				}
				dtSummary.WriteString(fmt.Sprintf("%s(%.0f), ", seat.Name, seat.NetAmt/10000))
			}
			dtSummary.WriteString("\n")
		}
	}

	// 4. Create Prompt (Focus on Review/Summary)
	prompt := fmt.Sprintf(`You are an expert Stock Market Analyst.
Your task is to provide a comprehensive "Market Review" (复盘) for the A-share market on %s.

Here is the market data:

[Sector Performance]
%s

[Limit-Up (Sentiment) Data]
%s

[Dragon Tiger List (Hot Money)]
%s

Please analyze the data and generate a structured review in Chinese (Simplified).

IMPORTANT: The provided data (Limit-Up, Dragon Tiger) might be empty if the market is closed or API fails.
If any data section is empty, explicitly state that "No sufficient data available" for that part, and DO NOT HALLUCINATE or invent stock names.
If Limit-Up Data is empty, do not list any "Hot Stocks" unless they are from the Dragon Tiger List or Sectors.
If NO data is available at all, return a summary stating that market data is unavailable.

Structure:
1. **Market Summary (市场总览)**: A brief summary of today's market emotion and main themes.
2. **Sector Analysis (板块分析)**: Which sectors are strong? Is there a clear main line? Where is the money flowing?
3. **Sentiment Analysis (情绪分析)**: Analyze the limit-up pool. Is the sentiment heating up or cooling down? Are there high-space stocks (连板高度)?
4. **Hot Money Analysis (游资动向)**: Based on Dragon Tiger List, where are the active funds?
5. **Risks (风险提示)**: Any potential risks based on the data?
6. **Opportunities (明日机会)**: Based on today's rotation, what to look for tomorrow?

Output ONLY a JSON object with the following fields:
{
  "summary": "...",
  "sector_analysis": "...",
  "sentiment_analysis": "...",
  "key_risks": ["risk1", "risk2"],
  "opportunities": ["opp1", "opp2"]
}

Ensure the response is valid JSON. Do not include markdown formatting like `+"```json"+`.
`, date, sectorSummary.String(), limitUpSummary.String(), dtSummary.String())

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: prompt},
			},
		},
	}

	// 5. Generate
	resp, err := llmClient.GenerateContent(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to generate review: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no content generated")
	}

	content := resp.Choices[0].Content
	log.Printf("Raw Review Response: %s", content)

	// 6. Parse JSON
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var review ai.MarketReviewResponse
	if err := json.Unmarshal([]byte(content), &review); err != nil {
		log.Printf("Failed to parse review JSON: %v. Raw: %s", err, content)
		// Fallback: put everything in summary
		return &ai.MarketReviewResponse{
			Summary: content,
		}, nil
	}

	return &review, nil
}

func (p *LangChainProvider) AnalyzeMarket(ctx context.Context, sectors []*stock.SectorInfo, limitUps []*stock.LimitUpStock, dragonTigerList []*stock.DragonTigerItem, date string) (*ai.MarketAnalysisResponse, error) {
	// 1. Determine ModelConfig
	var cfg ModelConfig
	if p.fileConfig != nil {
		var ok bool
		cfg, ok = p.fileConfig.Models[string(p.fileConfig.CurrentProvider)]
		if ok {
			cfg.Provider = p.fileConfig.CurrentProvider
		} else {
			return nil, fmt.Errorf("provider not found")
		}
	} else {
		return nil, fmt.Errorf("no config found")
	}

	log.Printf("Using LLM Provider for Market Analysis: %s, Model: %s", cfg.Provider, cfg.ModelName)

	// 2. Create LLM
	llmClient, err := NewModel(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create llm: %w", err)
	}

	// 3. Prepare Data Context (Reuse same context logic as ReviewMarket)
	var sectorSummary strings.Builder
	sectorSummary.WriteString("Top Sectors:\n")
	for i, s := range sectors {
		if i >= 10 {
			break
		}
		sectorSummary.WriteString(fmt.Sprintf("- %s: +%.2f%% (Net Inflow: %.2f), Top Stock: %s\n", s.Name, s.ChangePercent, s.NetInflow, s.TopStockName))
	}

	var limitUpSummary strings.Builder
	limitUpSummary.WriteString(fmt.Sprintf("Limit Up Pool (Total: %d):\n", len(limitUps)))
	typeCount := make(map[string]int)
	for _, s := range limitUps {
		typeCount[s.LimitUpType]++
		limitUpSummary.WriteString(fmt.Sprintf("- %s: %s, %s, %s\n", s.Name, s.LimitUpType, s.Reason, s.ChangePercent))
	}

	var dtSummary strings.Builder
	dtSummary.WriteString(fmt.Sprintf("Dragon Tiger List (Top 5 Net Buy):\n"))
	for i, item := range dragonTigerList {
		if i >= 5 {
			break
		}
		dtSummary.WriteString(fmt.Sprintf("- %s: +%.2f%%, Net: %.1f Wan, Reason: %s\n", item.Name, item.ChangePercent, item.NetInflow/10000, item.Reason))
	}

	// 4. Create Prompt (Focus on Prediction/Opportunity/Risk)
	prompt := fmt.Sprintf(`You are an expert Stock Market Analyst.
Your task is to provide a "Pre-market Analysis" (盘前分析) for the A-share market, based on the provided data.

Here is the latest available market data (representing the most recent trading session, usually yesterday or last Friday):

[Sector Performance]
%s

[Limit-Up (Sentiment) Data]
%s

[Dragon Tiger List (Hot Money)]
%s

Please analyze the data and generate a structured analysis focused on OPPORTUNITIES and RISKS for the NEXT trading day (the upcoming opening).
Do not limit your analysis to describing the past; use the data to PREDICT the future trend.

IMPORTANT: The provided data (Limit-Up, Dragon Tiger) might be empty if the market is closed or API fails.
If any data section is empty, explicitly state that "No sufficient data available" for that part, and DO NOT HALLUCINATE or invent stock names.
If Limit-Up Data is empty, do not list any "Hot Stocks" unless they are from the Dragon Tiger List or Sectors.
If NO data is available at all, return a summary stating that market data is unavailable.

Structure:
1. **Hot Stocks (热门股票)**: Identify 3-5 stocks that are likely to be active tomorrow based on limit-up momentum or dragon tiger list funds.
2. **Recommended Stocks (推荐关注)**: Recommend 1-3 stocks with strong logic (e.g., sector resonance, hot money inflow). Provide brief reasons.
3. **Risks (风险提示)**: What should traders watch out for in the next session? (e.g., high-level divergence, sector rotation failure).
4. **Opportunities (机会展望)**: Which sectors or themes might lead tomorrow?
5. **Analysis Summary (分析总结)**: A concise overview of the strategy for tomorrow.

Output ONLY a JSON object with the following fields:
{
  "hot_stocks": ["stock1", "stock2"],
  "recommended_stocks": ["stock1 (Reason)", "stock2 (Reason)"],
  "risks": ["risk1", "risk2"],
  "opportunities": ["opp1", "opp2"],
  "analysis_summary": "..."
}

Ensure the response is valid JSON. Do not include markdown formatting like `+"```json"+`.
`, sectorSummary.String(), limitUpSummary.String(), dtSummary.String())

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: prompt},
			},
		},
	}

	// 5. Generate
	resp, err := llmClient.GenerateContent(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to generate analysis: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no content generated")
	}

	content := resp.Choices[0].Content
	log.Printf("Raw Analysis Response: %s", content)

	// 6. Parse JSON
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var analysis ai.MarketAnalysisResponse
	if err := json.Unmarshal([]byte(content), &analysis); err != nil {
		log.Printf("Failed to parse analysis JSON: %v. Raw: %s", err, content)
		return &ai.MarketAnalysisResponse{
			AnalysisSummary: content,
		}, nil
	}

	return &analysis, nil
}
