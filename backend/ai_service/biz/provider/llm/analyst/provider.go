package analyst

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"stock_assistant/backend/ai_service/biz/provider/llm/core"
	"stock_assistant/backend/ai_service/biz/provider/prompt"
	ai "stock_assistant/backend/ai_service/kitex_gen/ai"
	"stock_assistant/backend/ai_service/kitex_gen/stock"
	"stock_assistant/backend/common/langfuse"

	"github.com/tmc/langchaingo/llms"
)

type Provider struct {
	fileConfig *core.FileConfig
	lf         *langfuse.LangfuseManager
}

func NewProvider(fileConfig *core.FileConfig) *Provider {
	return &Provider{
		fileConfig: fileConfig,
		lf:         langfuse.GetLangfuse(),
	}
}

func (p *Provider) ReviewMarket(ctx context.Context, sectors []*stock.SectorInfo, limitUps []*stock.LimitUpStock, dragonTigerList []*stock.DragonTigerItem, date string) (*ai.MarketReviewResponse, error) {
	// 1. 确定模型配置
	var cfg core.ModelConfig
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

	if p.lf != nil {
		traceID := p.lf.CreateTrace(ctx, "MarketReview", map[string]interface{}{
			"date": date,
		})
		ctx = langfuse.WithTraceID(ctx, traceID)
	}

	// 2. 创建 LLM
	llmClient, err := core.NewModel(ctx, cfg)
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
	ctx = core.WithModelName(ctx, cfg.ModelName)
	ctx = core.WithGenerationName(ctx, "LLM-ReviewMarket")
	resp, err := core.GenerateContentWithRetry(ctx, llmClient, messages)
	if err != nil {
		return nil, fmt.Errorf("生成复盘失败: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("未生成内容")
	}

	content := resp.Choices[0].Content

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

func (p *Provider) AnalyzeMarket(ctx context.Context, sectors []*stock.SectorInfo, limitUps []*stock.LimitUpStock, dragonTigerList []*stock.DragonTigerItem, date string) (*ai.MarketAnalysisResponse, []string, error) {
	// 1. 确定模型配置
	var cfg core.ModelConfig
	if p.fileConfig != nil {
		var ok bool
		cfg, ok = p.fileConfig.Models[string(p.fileConfig.CurrentProvider)]
		if ok {
			cfg.Provider = p.fileConfig.CurrentProvider
		} else {
			return nil, nil, fmt.Errorf("未找到提供商")
		}
	} else {
		return nil, nil, fmt.Errorf("未找到配置")
	}

	log.Printf("使用 LLM 提供商进行市场分析: %s, 模型: %s", cfg.Provider, cfg.ModelName)

	if p.lf != nil {
		traceID := p.lf.CreateTrace(ctx, "MarketAnalysis", map[string]interface{}{
			"date": date,
		})
		ctx = langfuse.WithTraceID(ctx, traceID)
	}

	// 2. 创建 LLM
	llmClient, err := core.NewModel(ctx, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("创建 llm 失败: %w", err)
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
		return nil, nil, fmt.Errorf("获取市场分析提示词失败")
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
	ctx = core.WithModelName(ctx, cfg.ModelName)
	ctx = core.WithGenerationName(ctx, "LLM-AnalyzeMarket")
	resp, err := core.GenerateContentWithRetry(ctx, llmClient, messages)
	if err != nil {
		return nil, nil, fmt.Errorf("生成分析失败: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, nil, fmt.Errorf("未生成内容")
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
		// 即使解析失败，尝试创建一个包含原始内容的响应
		analysis.AnalysisSummary = content
	} else {
		// 映射临时结构体到 IDL 结构体
		analysis.HotStocks = tempAnalysis.HotSectors

		var recStocks []string
		var stockCodes []string
		for _, s := range tempAnalysis.RecommendedStocks {
			recStocks = append(recStocks, fmt.Sprintf("%s (%s): %s", s.Name, s.Code, s.Reason))
			stockCodes = append(stockCodes, s.Code)
		}
		analysis.RecommendedStocks = recStocks

		analysis.Risks = tempAnalysis.Risks
		analysis.Opportunities = tempAnalysis.Opportunities
		analysis.AnalysisSummary = tempAnalysis.AnalysisSummary
		analysis.PolicyScore = tempAnalysis.PolicyScore
	}

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

	// 注意：自动追踪推荐股票逻辑暂时移除或移到 Handler 层，因为这里不持有 stockClient
	// 我们需要从 analysis 中提取 stockCodes (如果上面解析失败则为空)
	var stockCodes []string
	if len(tempAnalysis.RecommendedStocks) > 0 {
		for _, s := range tempAnalysis.RecommendedStocks {
			stockCodes = append(stockCodes, s.Code)
		}
	}

	return &analysis, stockCodes, nil
}

func (p *Provider) ProcessMarketTrends(ctx context.Context, items []*ai.RawTrendItem) ([]*ai.AnalyzedTrendItem, error) {
	if len(items) == 0 {
		return nil, nil
	}

	// 1. 确定模型配置
	var cfg core.ModelConfig
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

	if p.lf != nil {
		traceID := p.lf.CreateTrace(ctx, "MarketTrendsProcessing", map[string]interface{}{
			"items": len(items),
		})
		ctx = langfuse.WithTraceID(ctx, traceID)
	}

	// 2. 创建 LLM
	llmClient, err := core.NewModel(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("创建 llm 失败: %w", err)
	}

	// 3. 准备 JSON 输入
	type RawInput struct {
		Title   string `json:"title"`
		Source  string `json:"source"`
		Content string `json:"content"`
	}
	var inputs []RawInput
	for _, item := range items {
		inputs = append(inputs, RawInput{
			Title:   item.Title,
			Source:  item.Source,
			Content: item.Content,
		})
	}
	inputJSON, _ := json.Marshal(inputs)

	// 4. 创建提示词
	promptTemplate := prompt.GetManager().GetPrompt(ctx, prompt.MarketTrendsProcessing)
	if promptTemplate == "" {
		return nil, fmt.Errorf("获取市场信息处理提示词失败")
	}
	// 将输入数据附在提示词后面
	promptStr := fmt.Sprintf("%s\n\n输入数据:\n%s", promptTemplate, string(inputJSON))

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: promptStr},
			},
		},
	}

	// 5. 生成
	// 增加超时，因为处理可能较慢
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	ctxWithTimeout = core.WithModelName(ctxWithTimeout, cfg.ModelName)
	ctxWithTimeout = core.WithGenerationName(ctxWithTimeout, "LLM-ProcessTrends")
	resp, err := core.GenerateContentWithRetry(ctxWithTimeout, llmClient, messages)
	if err != nil {
		return nil, fmt.Errorf("生成分析失败: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("未生成内容")
	}

	content := resp.Choices[0].Content

	// 6. 解析 JSON
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// 定义中间结构以解析 JSON
	type AnalyzedOutput struct {
		Title              string   `json:"title"`
		Source             string   `json:"source"`
		Summary            string   `json:"summary"`
		FinancialRelevance int      `json:"financial_relevance"`
		RelatedSectors     []string `json:"related_sectors"`
		RelatedStocks      []string `json:"related_stocks"`
		ImpactAnalysis     string   `json:"impact_analysis"`
		SentimentScore     float64  `json:"sentiment_score"`
		ImpactType         string   `json:"impact_type"`
		ImpactScope        string   `json:"impact_scope"`
		Weight             float64  `json:"weight"`
	}
	var outputs []AnalyzedOutput
	if err := json.Unmarshal([]byte(content), &outputs); err != nil {
		log.Printf("解析趋势分析 JSON 失败: %v", err)
		return nil, err
	}

	// 7. 映射回结果
	// 需要根据 Title/Source 匹配原始 URL (因为 LLM 可能不输出 URL)
	// 假设顺序一致，或者通过 Map 匹配
	itemMap := make(map[string]*ai.RawTrendItem)
	for _, item := range items {
		key := item.Title + "|" + item.Source
		itemMap[key] = item
	}

	var results []*ai.AnalyzedTrendItem
	for _, out := range outputs {
		// Filter out items with no sectors, no stocks, and not market-wide scope
		if len(out.RelatedSectors) == 0 && len(out.RelatedStocks) == 0 && out.ImpactScope != "market_wide" {
			continue
		}

		// 尝试找回 URL
		url := ""
		if original, ok := itemMap[out.Title+"|"+out.Source]; ok {
			url = original.Url
		}

		results = append(results, &ai.AnalyzedTrendItem{
			Title:              out.Title,
			Source:             out.Source,
			Url:                url,
			Summary:            out.Summary,
			FinancialRelevance: int32(out.FinancialRelevance),
			RelatedSectors:     out.RelatedSectors,
			RelatedStocks:      out.RelatedStocks,
			SentimentScore:     out.SentimentScore,
			ImpactType:         out.ImpactType,
			ImpactScope:        out.ImpactScope,
			Weight:             out.Weight,
		})
	}

	return results, nil
}
