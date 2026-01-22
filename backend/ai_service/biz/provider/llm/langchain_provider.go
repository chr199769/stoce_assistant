package llm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"stock_assistant/backend/ai_service/biz/provider/langfuse"
	"stock_assistant/backend/ai_service/biz/provider/prompt"
	"stock_assistant/backend/ai_service/biz/tool"
	"stock_assistant/backend/ai_service/biz/tool/fractal"
	ai "stock_assistant/backend/ai_service/kitex_gen/ai"
	"stock_assistant/backend/ai_service/kitex_gen/stock"
	"stock_assistant/backend/ai_service/kitex_gen/stock/stockservice"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
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

func (p *LangChainProvider) Predict(ctx context.Context, stockCode string, days int32, modelName string) (string, float64, string, string, error) {
	if modelName == "fractal" {
		res, conf, summary, err := p.predictWithFractal(ctx, stockCode, days)
		return res, conf, summary, "", err
	}

	// 1. 确定模型配置
	var cfg ModelConfig
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
					cfg.Provider = ModelProvider(provider)
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
				return "", 0, "", "", fmt.Errorf("未找到提供商且禁用了 fake 提供商")
			}
		}
	} else {
		// 无文件配置，回退
		return "", 0, "", "", fmt.Errorf("未找到配置且禁用了 fake 提供商")
	}

	log.Printf("使用 LLM 提供商: %s, 模型: %s", cfg.Provider, cfg.ModelName)

	// 2. 创建 LLM
	llm, err := NewModel(ctx, cfg)
	if err != nil {
		return "", 0, "", "", fmt.Errorf("创建 llm 失败: %w", err)
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
	stockTool := tool.NewStockPriceTool(p.stockClient)
	marketTool := tool.NewMarketInfoTool()
	analysisTool := tool.NewStockAnalysisTool()
	sectorTool := tool.NewSectorTool(p.stockClient)
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
		return "", 0, "", "", fmt.Errorf("获取预测用户提示词失败")
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
	resp, err := llm.GenerateContent(ctx, messages)
	if err != nil {
		return "", 0, "", "", fmt.Errorf("llm 生成失败: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", 0, "", "", fmt.Errorf("llm 响应为空")
	}
	res := resp.Choices[0].Content

	// 解析输出
	analysis := res
	confidence := 0.5 // 默认值
	newsSummary := "详见分析"

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
			Confidence  float64 `json:"confidence"`
			NewsSummary string  `json:"news_summary"`
		}
		if err := json.Unmarshal([]byte(metadataJSON), &metadata); err == nil {
			confidence = metadata.Confidence
			newsSummary = metadata.NewsSummary

			// 用于 Langfuse
			metadataMap = map[string]interface{}{
				"confidence":   confidence,
				"news_summary": newsSummary,
			}
		} else {
			log.Printf("解析元数据 JSON 失败: %v. JSON: %s", err, metadataJSON)
		}
	} else {
		log.Printf("响应中未找到元数据分隔符")
	}

	// 使用 Langfuse 追踪
	if lf != nil {
		lf.CreateGeneration(ctx, traceID, "LLM-Predict", cfg.ModelName, userPrompt, analysis, metadataMap, time.Now(), time.Now())
		lf.UpdateTrace(ctx, traceID, userPrompt, analysis)
	}

	return analysis, confidence, newsSummary, traceID, nil
}

func (p *LangChainProvider) RecognizeImage(ctx context.Context, imageData []byte, modelName string) ([]*ai.RecognizedStock, error) {
	// 1. 确定模型配置
	var cfg ModelConfig
	if p.fileConfig != nil {
		// 在所有提供商中查找模型配置
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
				return nil, fmt.Errorf("未找到提供商且禁用了 fake 提供商")
			}
		}
	} else {
		return nil, fmt.Errorf("未找到配置且禁用了 fake 提供商")
	}

	log.Printf("使用 LLM 提供商进行图像识别: %s, 模型: %s", cfg.Provider, cfg.ModelName)

	// 2. 创建 LLM
	llmClient, err := NewModel(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("创建 llm 失败: %w", err)
	}

	// 3. 准备图像
	// 检测内容类型
	mimeType := http.DetectContentType(imageData)
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	imageURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Image)

	// 4. 创建消息
	promptStr := prompt.GetManager().GetPrompt(ctx, prompt.ImageRecognitionMaster)
	if promptStr == "" {
		return nil, fmt.Errorf("获取图像识别提示词失败")
	}

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: promptStr},
				llms.ImageURLContent{URL: imageURL},
			},
		},
	}

	// 5. 生成内容
	resp, err := llmClient.GenerateContent(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("生成内容失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("未生成内容")
	}

	content := resp.Choices[0].Content
	log.Printf("LLM 原始响应: %s", content)

	// 6. 解析 JSON
	// 清理可能的 markdown 代码块
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var stocks []*ai.RecognizedStock
	// 使用临时结构体解析以处理可能的字段不匹配
	var tempStocks []struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}

	if err := json.Unmarshal([]byte(content), &tempStocks); err != nil {
		// 如果 JSON 解析失败，尝试使用正则
		log.Printf("JSON 解析失败: %v, 尝试使用正则", err)
		// 简单正则匹配 sh/sz + 6位数字 或 纯6位数字
		re := regexp.MustCompile(`((sh|sz)\d{6})|(\d{6})`)
		matches := re.FindAllString(content, -1)

		uniqueCodes := make(map[string]bool)

		for _, match := range matches {
			code := match
			// 如果缺少前缀则修复
			if len(code) == 6 && !strings.HasPrefix(code, "sh") && !strings.HasPrefix(code, "sz") {
				if strings.HasPrefix(code, "6") {
					code = "sh" + code
				} else if strings.HasPrefix(code, "0") || strings.HasPrefix(code, "3") {
					code = "sz" + code
				}
				// 忽略其他情况
			}

			// 避免重复并确保格式有效
			if !uniqueCodes[code] && (strings.HasPrefix(code, "sh") || strings.HasPrefix(code, "sz")) {
				uniqueCodes[code] = true
				stocks = append(stocks, &ai.RecognizedStock{
					Code: code,
					Name: "未知", // 无法通过正则可靠提取名称
				})
			}
		}

		if len(stocks) == 0 {
			// 如果正则也未找到任何内容，返回空列表而不是错误
			log.Printf("从图像解析股票信息失败 (JSON 和正则均失败): %v", err)
			return []*ai.RecognizedStock{}, nil
		}
	} else {
		for _, s := range tempStocks {
			code := s.Code
			// 如果缺少前缀则修复
			if len(code) == 6 && !strings.HasPrefix(code, "sh") && !strings.HasPrefix(code, "sz") {
				if strings.HasPrefix(code, "6") {
					code = "sh" + code
				} else if strings.HasPrefix(code, "0") || strings.HasPrefix(code, "3") {
					code = "sz" + code
				}
				// 忽略其他情况
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
	// ... (市场复盘实现 - 关注今日总结)
	// 1. 确定模型配置
	var cfg ModelConfig
	if p.fileConfig != nil {
		var ok bool
		cfg, ok = p.fileConfig.Models[string(p.fileConfig.CurrentProvider)]
		if ok {
			cfg.Provider = p.fileConfig.CurrentProvider
		} else {
			return nil, fmt.Errorf("未找到提供商")
		}
	} else {
		return nil, fmt.Errorf("未找到配置")
	}

	log.Printf("使用 LLM 提供商进行市场复盘: %s, 模型: %s", cfg.Provider, cfg.ModelName)

	// 2. 创建 LLM
	llmClient, err := NewModel(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("创建 llm 失败: %w", err)
	}

	// 3. 准备数据上下文
	var sectorSummary strings.Builder
	sectorSummary.WriteString("热门板块:\n")
	for i, s := range sectors {
		if i >= 10 { // 前10
			break
		}
		sectorSummary.WriteString(fmt.Sprintf("- %s: +%.2f%% (净流入: %.2f), 领涨股: %s\n", s.Name, s.ChangePercent, s.NetInflow, s.TopStockName))
	}

	var limitUpSummary strings.Builder
	limitUpSummary.WriteString(fmt.Sprintf("涨停池 (共: %d):\n", len(limitUps)))
	// 简单统计
	typeCount := make(map[string]int)
	for _, s := range limitUps {
		typeCount[s.LimitUpType]++
		limitUpSummary.WriteString(fmt.Sprintf("- %s: %s, %s, %.2f%%\n", s.Name, s.LimitUpType, s.Reason, s.ChangePercent))
	}

	var dtSummary strings.Builder
	dtSummary.WriteString(fmt.Sprintf("龙虎榜 (净买入前5):\n"))
	for i, item := range dragonTigerList {
		if i >= 5 {
			break
		}
		dtSummary.WriteString(fmt.Sprintf("- %s: +%.2f%%, 净额: %.1f 万, 原因: %s\n", item.Name, item.ChangePercent, item.NetInflow/10000, item.Reason))
		// 添加席位 (前3)
		if len(item.BuySeats) > 0 {
			dtSummary.WriteString("  [买入席位]: ")
			for k, seat := range item.BuySeats {
				if k >= 2 {
					break
				}
				dtSummary.WriteString(fmt.Sprintf("%s(%.0f), ", seat.Name, seat.NetAmt/10000))
			}
			dtSummary.WriteString("\n")
		}
	}

	// 4. 创建提示词 (关注复盘/总结)
	promptTemplate := prompt.GetManager().GetPrompt(ctx, prompt.MarketReviewMaster)
	if promptTemplate == "" {
		return nil, fmt.Errorf("获取市场复盘提示词失败")
	}
	promptStr := fmt.Sprintf(promptTemplate, date, sectorSummary.String(), limitUpSummary.String(), dtSummary.String())

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: promptStr},
			},
		},
	}

	// 5. 生成
	resp, err := llmClient.GenerateContent(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("生成复盘失败: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("未生成内容")
	}

	content := resp.Choices[0].Content
	log.Printf("原始复盘响应: %s", content)

	// 6. 解析 JSON
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var review ai.MarketReviewResponse
	if err := json.Unmarshal([]byte(content), &review); err != nil {
		log.Printf("解析复盘 JSON 失败: %v. 原始内容: %s", err, content)
		// 回退: 将所有内容放入 summary
		return &ai.MarketReviewResponse{
			Summary: content,
		}, nil
	}

	return &review, nil
}

func (p *LangChainProvider) AnalyzeMarket(ctx context.Context, sectors []*stock.SectorInfo, limitUps []*stock.LimitUpStock, dragonTigerList []*stock.DragonTigerItem, date string) (*ai.MarketAnalysisResponse, error) {
	// 1. 确定模型配置
	var cfg ModelConfig
	if p.fileConfig != nil {
		var ok bool
		cfg, ok = p.fileConfig.Models[string(p.fileConfig.CurrentProvider)]
		if ok {
			cfg.Provider = p.fileConfig.CurrentProvider
		} else {
			return nil, fmt.Errorf("未找到提供商")
		}
	} else {
		return nil, fmt.Errorf("未找到配置")
	}

	log.Printf("使用 LLM 提供商进行市场分析: %s, 模型: %s", cfg.Provider, cfg.ModelName)

	// 2. 创建 LLM
	llmClient, err := NewModel(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("创建 llm 失败: %w", err)
	}

	// 3. 准备数据上下文 (复用 ReviewMarket 的逻辑)
	var sectorSummary strings.Builder
	sectorSummary.WriteString("热门板块:\n")
	for i, s := range sectors {
		if i >= 10 {
			break
		}
		sectorSummary.WriteString(fmt.Sprintf("- %s: +%.2f%% (净流入: %.2f), 领涨股: %s\n", s.Name, s.ChangePercent, s.NetInflow, s.TopStockName))
	}

	var limitUpSummary strings.Builder
	limitUpSummary.WriteString(fmt.Sprintf("涨停池 (共: %d):\n", len(limitUps)))
	typeCount := make(map[string]int)
	for _, s := range limitUps {
		typeCount[s.LimitUpType]++
		limitUpSummary.WriteString(fmt.Sprintf("- %s: %s, %s, %.2f%%\n", s.Name, s.LimitUpType, s.Reason, s.ChangePercent))
	}

	var dtSummary strings.Builder
	dtSummary.WriteString(fmt.Sprintf("龙虎榜 (净买入前5):\n"))
	for i, item := range dragonTigerList {
		if i >= 5 {
			break
		}
		dtSummary.WriteString(fmt.Sprintf("- %s: +%.2f%%, 净额: %.1f 万, 原因: %s\n", item.Name, item.ChangePercent, item.NetInflow/10000, item.Reason))
	}

	// 4. 创建提示词 (关注预测/机会/风险)
	promptTemplate := prompt.GetManager().GetPrompt(ctx, prompt.MarketAnalysisMaster)
	if promptTemplate == "" {
		return nil, fmt.Errorf("获取市场分析提示词失败")
	}
	promptStr := fmt.Sprintf(promptTemplate, sectorSummary.String(), limitUpSummary.String(), dtSummary.String())

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: promptStr},
			},
		},
	}

	// 5. 生成
	resp, err := llmClient.GenerateContent(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("生成分析失败: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("未生成内容")
	}

	content := resp.Choices[0].Content
	log.Printf("原始分析响应: %s", content)

	// 6. 解析 JSON
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var analysis ai.MarketAnalysisResponse

	// 定义临时结构体以匹配新提示词 JSON 结构但映射到旧 IDL 以兼容
	type MarketAnalysisTemp struct {
		HotSectors        []string `json:"hot_sectors"`
		RecommendedStocks []struct {
			Code   string `json:"code"`
			Name   string `json:"name"`
			Reason string `json:"reason"`
		} `json:"recommended_stocks"`
		Risks           []string `json:"risks"`
		Opportunities   []string `json:"opportunities"`
		AnalysisSummary string   `json:"analysis_summary"`
		PolicyScore     float64  `json:"policy_score"`
	}

	var tempAnalysis MarketAnalysisTemp
	if err := json.Unmarshal([]byte(content), &tempAnalysis); err != nil {
		log.Printf("解析分析 JSON 失败: %v. 原始内容: %s", err, content)
		return &ai.MarketAnalysisResponse{
			AnalysisSummary: content,
		}, nil
	}

	// 映射临时结构体到 IDL 结构体
	analysis.HotStocks = tempAnalysis.HotSectors

	var recStocks []string
	for _, s := range tempAnalysis.RecommendedStocks {
		recStocks = append(recStocks, fmt.Sprintf("%s (%s): %s", s.Name, s.Code, s.Reason))
	}
	analysis.RecommendedStocks = recStocks

	analysis.Risks = tempAnalysis.Risks
	analysis.Opportunities = tempAnalysis.Opportunities
	analysis.AnalysisSummary = tempAnalysis.AnalysisSummary
	analysis.PolicyScore = tempAnalysis.PolicyScore

	// 7. 计算情绪得分 (确定性算法)
	// 基准: 50
	// 因子 1: 涨停数 (0-30 -> 0-30 分)
	// 因子 2: 炸板 (负面影响)
	// 因子 3: 净流入 (热门板块)

	sentimentScore := 50.0
	limitUpCount := float64(len(limitUps))
	if limitUpCount > 60 {
		limitUpCount = 60
	} // 上限
	sentimentScore += limitUpCount * 0.5

	// 检查炸板
	brokenCount := 0
	for _, s := range limitUps {
		if s.IsBroken {
			brokenCount++
		}
	}
	sentimentScore -= float64(brokenCount) * 1.0

	// 板块流入 (前5总和)
	inflowSum := 0.0
	for i, s := range sectors {
		if i >= 5 {
			break
		}
		inflowSum += s.NetInflow
	}
	// 归一化流入 (例如, 1亿 -> 1分, 最大 10分)
	// 假设单位是万, 所以 10000万 = 1亿.
	// 假设 50亿流入非常好.
	// 50亿 = 500,000万.
	inflowScore := inflowSum / 50000.0
	if inflowScore > 10 {
		inflowScore = 10
	}
	if inflowScore < -10 {
		inflowScore = -10
	}
	sentimentScore += inflowScore

	// 限制 0-100
	if sentimentScore > 100 {
		sentimentScore = 100
	}
	if sentimentScore < 0 {
		sentimentScore = 0
	}

	analysis.SentimentScore = sentimentScore

	// 8. 自动追踪推荐股票
	// 为每只推荐股票触发预测生成
	go func() {
		stocks := tempAnalysis.RecommendedStocks
		for _, s := range stocks {
			// 使用带超时的 detached context
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			name, code := s.Name, s.Code

			log.Printf("[自动追踪] 为推荐股票生成预测: %s (%s)", name, code)
			analysis, confidence, _, traceID, err := p.Predict(ctx, code, 3, "glm-4.6v-flash") // 使用智谱模型
			if err != nil {
				log.Printf("[自动追踪] 生成预测失败 %s: %v", code, err)
				continue
			}

			// 保存到数据库
			trend := "中性"
			if strings.Contains(analysis, "看涨") || strings.Contains(analysis, "Up") {
				trend = "看涨"
			} else if strings.Contains(analysis, "看跌") || strings.Contains(analysis, "Down") {
				trend = "看跌"
			}

			saveReq := &stock.SavePredictionRequest{
				Record: &stock.PredictionRecord{
					Id:             fmt.Sprintf("auto-%d-%s", time.Now().Unix(), code), // 简单ID
					StockCode:      code,
					PredictionDate: time.Now().Format("2006-01-02 15:04:05"),
					Content:        analysis,
					Confidence:     confidence,
					Trend:          trend,
					TraceId:        traceID,
				},
			}

			_, saveErr := p.stockClient.SavePrediction(ctx, saveReq)
			if saveErr != nil {
				log.Printf("[自动追踪] 保存预测失败 %s: %v", code, saveErr)
			} else {
				log.Printf("[自动追踪] 成功追踪 %s", code)
			}
			time.Sleep(time.Duration(1000) * time.Millisecond)
		}
	}()

	return &analysis, nil
}

func (p *LangChainProvider) predictWithFractal(ctx context.Context, stockCode string, days int32) (string, float64, string, error) {
	// 1. 获取历史 K 线 (例如 500 天)
	req := &stock.GetHistoricalKlineRequest{
		StockCode: stockCode,
		Days:      500, // 足够的历史用于匹配
	}
	resp, err := p.stockClient.GetHistoricalKline(ctx, req)
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
