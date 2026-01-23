package predictor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"stock_assistant/backend/ai_service/biz/provider/langfuse"
	"stock_assistant/backend/ai_service/biz/provider/llm/core"
	"stock_assistant/backend/ai_service/biz/provider/prompt"
	"stock_assistant/backend/ai_service/biz/rpc"
	"stock_assistant/backend/ai_service/biz/tool"
	"stock_assistant/backend/ai_service/biz/tool/fractal"
	"stock_assistant/backend/ai_service/kitex_gen/stock"

	"github.com/tmc/langchaingo/llms"
)

type Provider struct {
	fileConfig *core.FileConfig
}

func NewProvider(fileConfig *core.FileConfig) *Provider {
	return &Provider{
		fileConfig: fileConfig,
	}
}

// IsTradingTime 检查当前是否在A股交易时段 (周一至周五 9:15-15:00)
// 这是一个简化检查，未考虑法定节假日。
func IsTradingTime() bool {
	// 强制使用上海时区
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		// 加载失败则回退到固定时区 (CST)
		loc = time.FixedZone("CST", 8*3600)
	}

	now := time.Now().In(loc)
	weekday := now.Weekday()

	// 1. 检查是否为周末
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}

	// 2. 检查时间范围 09:15 - 15:00
	// 转换为分钟数以便比较
	currentMinutes := now.Hour()*60 + now.Minute()
	startMinutes := 9*60 + 15 // 09:15
	endMinutes := 15*60 + 0   // 15:00

	return currentMinutes >= startMinutes && currentMinutes <= endMinutes
}

func (p *Provider) Predict(ctx context.Context, stockCode string, days int32, modelName string) (string, float64, string, string, float64, string, error) {
	if modelName == "fractal" {
		res, conf, summary, err := p.predictWithFractal(ctx, stockCode, days)
		return res, conf, summary, "", 0, "", err
	}

	// 1. 确定模型配置
	var cfg core.ModelConfig
	if p.fileConfig != nil {
		// 在所有提供商中查找模型配置
		found := false
		// 如果提供了模型名称，则进行搜索
		if modelName != "" {
			log.Printf("正在搜索模型: %s", modelName)
			for provider, c := range p.fileConfig.Models {
				log.Printf("检查提供商: %s, 模型: %s", provider, c.ModelName)
				if c.ModelName == modelName {
					cfg = c
					cfg.Provider = core.ModelProvider(provider)
					found = true
					break
				}
			}
		}

		// 如果未找到或未提供，使用当前提供商
		if !found {
			if modelName != "" {
				log.Printf("配置中未找到模型 %s，回退到当前提供商", modelName)
			}

			var ok bool
			cfg, ok = p.fileConfig.Models[string(p.fileConfig.CurrentProvider)]
			if ok {
				cfg.Provider = p.fileConfig.CurrentProvider
			} else {
				return "", 0, "", "", 0, "", fmt.Errorf("未找到提供商且禁用了 fake 提供商")
			}
		}
	} else {
		// 无文件配置，回退
		return "", 0, "", "", 0, "", fmt.Errorf("未找到配置且禁用了 fake 提供商")
	}

	log.Printf("使用 LLM 提供商: %s, 模型: %s", cfg.Provider, cfg.ModelName)

	// 2. 创建 LLM
	llm, err := core.NewModel(ctx, cfg)
	if err != nil {
		return "", 0, "", "", 0, "", fmt.Errorf("创建 llm 失败: %w", err)
	}

	// 启动追踪
	var traceID string
	var lf *langfuse.LangfuseManager
	if lf = langfuse.GetLangfuse(); lf != nil {
		traceID = lf.CreateTrace(ctx, "StockPrediction", map[string]interface{}{
			"stock_code": stockCode,
			"env":        "dev",
		})
	}

	// 3. 创建工具 (用于预取数据，不传递给 LLM)
	stockTool := tool.NewStockPriceTool(rpc.StockClient)
	marketTool := tool.NewMarketInfoTool()
	analysisTool := tool.NewStockAnalysisTool()
	sectorTool := tool.NewSectorTool(rpc.StockClient)
	dtTool := tool.NewDragonTigerTool()

	// 预取股票数据
	var stockData string
	var errFetch error
	if lf != nil {
		start := time.Now()
		stockData, errFetch = stockTool.Call(ctx, stockCode)
		lf.Span(ctx, traceID, nil, "Tool:StockPrice", stockCode, stockData, start, time.Now())
	} else {
		stockData, errFetch = stockTool.Call(ctx, stockCode)
	}
	if errFetch != nil {
		log.Printf("预取股票数据失败: %v", errFetch)
		stockData = fmt.Sprintf("获取股票数据出错: %v", errFetch)
	}

	// 预取市场信息
	var marketInfo string
	if lf != nil {
		start := time.Now()
		marketInfo, _ = marketTool.Call(ctx, stockCode)
		lf.Span(ctx, traceID, nil, "Tool:MarketInfo", stockCode, marketInfo, start, time.Now())
	} else {
		marketInfo, _ = marketTool.Call(ctx, stockCode)
	}

	var analysisData string
	if lf != nil {
		start := time.Now()
		analysisData, _ = analysisTool.Call(ctx, stockCode)
		lf.Span(ctx, traceID, nil, "Tool:Analysis", stockCode, analysisData, start, time.Now())
	} else {
		analysisData, _ = analysisTool.Call(ctx, stockCode)
	}

	// 预取市场趋势 (已集成到 MarketInfoTool)
	// var marketTrendsContext string

	// 预取宏观背景
	var sectorContext string
	if lf != nil {
		start := time.Now()
		sectorContext, _ = sectorTool.Call(ctx, "industry")
		lf.Span(ctx, traceID, nil, "Tool:Sector", "industry", sectorContext, start, time.Now())
	} else {
		sectorContext, _ = sectorTool.Call(ctx, "industry")
	}

	var dtContext string
	if lf != nil {
		start := time.Now()
		dtContext, _ = dtTool.Call(ctx, "")
		lf.Span(ctx, traceID, nil, "Tool:DragonTiger", "today", dtContext, start, time.Now())
	} else {
		dtContext, _ = dtTool.Call(ctx, "")
	}

	// 预取分形分析 (仅供参考)
	var fractalAnalysisContext string
	startFractal := time.Now()
	fractalRes, _, _, err := p.predictWithFractal(ctx, stockCode, days)
	if err == nil {
		// 清理分形响应中的元数据，保持提示词整洁
		parts := strings.Split(fractalRes, "---METADATA---")
		if len(parts) > 0 {
			fractalAnalysisContext = strings.TrimSpace(parts[0])
		}
	} else {
		fractalAnalysisContext = "分形分析不可用。"
	}
	if lf != nil {
		lf.Span(ctx, traceID, nil, "Tool:Fractal", stockCode, fractalAnalysisContext, startFractal, time.Now())
	}

	// 4. 构建消息 (系统 + 用户)
	// 确定交易状态和上下文
	isTrading := IsTradingTime()
	tradingStatusStr := "已收盘"
	predictionFocus := "次日及未来3日预测"
	timeContextInstruction := `
- 当前状态：已收盘（盘后/周末）
- 重点：总结全天表现，分析龙虎榜数据，并提供下一个交易日及未来3日的展望。
- 盘口数据相关性：低（收盘后快照数据相关性较低）。
`

	if isTrading {
		tradingStatusStr = "盘中交易 (9:15-15:00)"
		predictionFocus = "当日收盘及未来3日预测"
		timeContextInstruction = `
- 当前状态：盘中交易（实时市场）
- 重点：分析实时盘口压力（总买/卖量）、委比/委差以及即时动能。
- 盘口数据相关性：高。使用它来预测今日剩余时间的股价走势。
`
	}

	// 获取系统提示词
	systemPromptTemplate := prompt.GetManager().GetPrompt(ctx, prompt.StockPredictionSystem)
	systemPrompt := fmt.Sprintf(systemPromptTemplate, timeContextInstruction, time.Now().Format("2006-01-02 15:04:05"), predictionFocus)

	// 获取用户提示词
	userPromptTemplate := prompt.GetManager().GetPrompt(ctx, prompt.StockPredictionUser)
	if userPromptTemplate == "" {
		return "", 0, "", "", 0, "", fmt.Errorf("获取预测用户提示词失败")
	}

	// 格式化用户提示词
	// 占位符: stockCode, time, time, stockData, analysisData, marketInfo, sectorContext, dtContext, fractal
	userPrompt := fmt.Sprintf(userPromptTemplate,
		stockCode,
		time.Now().Format("2006-01-02 15:04:05"),
		tradingStatusStr,
		stockData,
		analysisData,
		marketInfo,
		sectorContext,
		dtContext,
		fractalAnalysisContext,
	)

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: systemPrompt}},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: userPrompt}},
		},
	}

	// 5. 运行 LLM
	genStart := time.Now()
	resp, err := llm.GenerateContent(ctx, messages)
	genEnd := time.Now()
	if err != nil {
		return "", 0, "", "", 0, "", fmt.Errorf("llm 生成失败: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", 0, "", "", 0, "", fmt.Errorf("llm 响应为空")
	}
	res := resp.Choices[0].Content

	// 解析输出
	analysis := res
	confidence := 0.5 // 默认值
	newsSummary := "详见分析"
	policyImpactScope := "specific" // 默认值
	predictedChange := 0.0          // 默认值

	parts := strings.Split(res, "---METADATA---")
	var metadataMap map[string]interface{}

	if len(parts) > 1 {
		analysis = strings.TrimSpace(parts[0])
		metadataJSON := strings.TrimSpace(parts[1])
		// 清理 JSON (移除可能的 markdown 代码块)
		metadataJSON = strings.TrimPrefix(metadataJSON, "```json")
		metadataJSON = strings.TrimPrefix(metadataJSON, "```")
		metadataJSON = strings.TrimSuffix(metadataJSON, "```")
		metadataJSON = strings.TrimSpace(metadataJSON)

		var metadata struct {
			Confidence        float64 `json:"confidence"`
			NewsSummary       string  `json:"news_summary"`
			PolicyImpactScope string  `json:"policy_impact_scope"`
			PredictedChange   float64 `json:"predicted_change"`
		}
		if err := json.Unmarshal([]byte(metadataJSON), &metadata); err == nil {
			confidence = metadata.Confidence
			newsSummary = metadata.NewsSummary
			if metadata.PolicyImpactScope != "" {
				policyImpactScope = metadata.PolicyImpactScope
			}
			predictedChange = metadata.PredictedChange

			// 用于 Langfuse
			metadataMap = map[string]interface{}{
				"confidence":          confidence,
				"news_summary":        newsSummary,
				"policy_impact_scope": policyImpactScope,
				"predicted_change":    predictedChange,
			}
		} else {
			log.Printf("解析元数据 JSON 失败: %v. JSON: %s", err, metadataJSON)
		}
	} else {
		log.Printf("响应中未找到元数据分隔符")
	}

	// 使用 Langfuse 追踪
	if lf != nil {
		// 转换消息格式以适配 Langfuse (OpenAI 风格)
		var lfInput []map[string]interface{}
		for _, msg := range messages {
			var contentStr string
			for _, part := range msg.Parts {
				if textPart, ok := part.(llms.TextContent); ok {
					contentStr += textPart.Text
				} else if imgPart, ok := part.(llms.ImageURLContent); ok {
					contentStr += fmt.Sprintf("[Image: %s]", imgPart.URL)
				}
			}

			// 映射 role: human->user, ai->assistant
			roleStr := strings.ToLower(string(msg.Role))
			if roleStr == "human" {
				roleStr = "user"
			} else if roleStr == "ai" {
				roleStr = "assistant"
			}

			lfInput = append(lfInput, map[string]interface{}{
				"role":    roleStr,
				"content": contentStr,
			})
		}

		// 将 Output 也封装为 Chat Message 格式，以便 Langfuse 更好展示
		lfOutput := map[string]interface{}{
			"role":    "assistant",
			"content": analysis,
		}

		lf.CreateGeneration(ctx, traceID, "LLM-Predict", cfg.ModelName, lfInput, lfOutput, metadataMap, genStart, genEnd)
		lf.UpdateTrace(ctx, traceID, userPrompt, analysis)
	}

	return analysis, confidence, newsSummary, traceID, predictedChange, policyImpactScope, nil
}

func (p *Provider) predictWithFractal(ctx context.Context, stockCode string, days int32) (string, float64, string, error) {
	// 1. 获取历史 K 线 (例如 500 天)
	req := &stock.GetHistoricalKlineRequest{
		StockCode: stockCode,
		Days:      500, // 足够的历史用于匹配
	}
	if rpc.StockClient == nil {
		return "", 0, "", fmt.Errorf("Stock Service 未初始化")
	}
	resp, err := rpc.StockClient.GetHistoricalKline(ctx, req)
	if err != nil {
		return "", 0, "", fmt.Errorf("获取K线失败: %v", err)
	}

	klines := resp.Klines
	if len(klines) < 60 { // 需要至少一些历史 + 查询模式
		return "历史数据不足，无法进行分形分析。", 0, "", nil
	}

	// 2. 准备数据
	// 使用收盘价进行匹配
	closes := make([]float64, len(klines))
	for i, k := range klines {
		closes[i] = k.Close
	}

	// 3. 定义查询模式 (最近 20 天)
	queryLen := 20
	if len(closes) < queryLen*2 {
		return "历史太短，无法进行模式匹配。", 0, "", nil
	}

	queryPattern := closes[len(closes)-queryLen:]
	searchSpace := closes[:len(closes)-queryLen] // 在过去搜索，排除当前模式

	// 4. 执行匹配 (皮尔逊相关系数)
	bestSim := -1.0
	var bestMatch []float64
	var bestMatchIdx int

	// 归一化查询
	normQuery := fractal.NormalizeSeries(queryPattern)

	for i := 0; i <= len(searchSpace)-queryLen-int(days); i++ {
		candidate := searchSpace[i : i+queryLen]
		normCandidate := fractal.NormalizeSeries(candidate)

		sim := fractal.CalculatePearsonCorrelation(normQuery, normCandidate)
		if sim > bestSim {
			bestSim = sim
			bestMatchIdx = i
			// 获取匹配后的未来 'days' 天价格
			bestMatch = searchSpace[i+queryLen : i+queryLen+int(days)]
		}
	}

	if bestMatch == nil {
		return "历史中未找到相似模式。", 0, "", nil
	}

	// 5. 生成推演
	// 计算最佳匹配未来的百分比变化
	// 并将其应用于当前价格。
	currentPrice := closes[len(closes)-1]
	projection := make([]float64, len(bestMatch))

	startPriceMatch := searchSpace[bestMatchIdx+queryLen-1]

	var analysisBuilder strings.Builder
	analysisBuilder.WriteString(fmt.Sprintf("发现分形模式匹配 (相似度: %.2f%%)\n", bestSim*100))
	analysisBuilder.WriteString(fmt.Sprintf("匹配历史时段: %s\n", klines[bestMatchIdx].Date)) // 近似日期
	analysisBuilder.WriteString("走势推演:\n")

	for i, price := range bestMatch {
		changeRatio := price / startPriceMatch
		projectedPrice := currentPrice * changeRatio
		projection[i] = projectedPrice
		analysisBuilder.WriteString(fmt.Sprintf("未来第 %d 天: %.2f\n", i+1, projectedPrice))
	}

	trend := "中性"
	if len(projection) > 0 {
		if projection[len(projection)-1] > currentPrice*1.02 {
			trend = "看涨"
		} else if projection[len(projection)-1] < currentPrice*0.98 {
			trend = "看跌"
		}
	}

	finalAnalysis := fmt.Sprintf("基于分形几何学分析，当前20日K线形态与 %s 开始的历史走势有 %.0f%% 的相似度。\n\n趋势预测: %s\n\n%s", klines[bestMatchIdx].Date, bestSim*100, trend, analysisBuilder.String())

	// 准备图表的分形数据
	fractalData := map[string]interface{}{
		"query":      queryPattern,
		"match":      bestMatch,
		"projection": projection,
		"match_date": klines[bestMatchIdx].Date,
		"dates":      make([]string, len(queryPattern)+len(projection)), // 需要时填充日期占位符
	}
	fractalDataJSON, _ := json.Marshal(fractalData)

	// 格式化元数据
	metadata := fmt.Sprintf(`---METADATA---
{"confidence": %.2f, "news_summary": "基于历史分形自相似性的技术分析。", "fractal_data": %s}`, bestSim, string(fractalDataJSON))

	return finalAnalysis + "\n\n" + metadata, bestSim, "分形模式匹配", nil
}
