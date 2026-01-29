package predictor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"stock_assistant/backend/ai_service/biz/provider/llm/core"
	"stock_assistant/backend/ai_service/biz/provider/prompt"
	"stock_assistant/backend/ai_service/biz/rpc"
	"stock_assistant/backend/ai_service/biz/tool"
	"stock_assistant/backend/ai_service/biz/tool/fractal"
	"stock_assistant/backend/ai_service/config"
	"stock_assistant/backend/ai_service/kitex_gen/stock"
	eastmoney "stock_assistant/backend/common/eastmoney"
	"stock_assistant/backend/common/langfuse"
	knowledge_graph "stock_assistant/backend/knowledge_graph/kitex_gen/knowledge_graph"

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

	cfg, err := p.resolveModelConfig(modelName)
	if err != nil {
		return "", 0, "", "", 0, "", err
	}

	log.Printf("使用 LLM 提供商: %s, 模型: %s", cfg.Provider, cfg.ModelName)

	llm, err := core.NewModel(ctx, cfg)
	if err != nil {
		return "", 0, "", "", 0, "", fmt.Errorf("创建 llm 失败: %w", err)
	}

	ctx, lf := startPredictionTrace(ctx, stockCode)
	data := p.collectPredictionContext(ctx, stockCode, days, lf)
	tradingStatusStr, predictionFocus, timeContextInstruction := buildTradingContext()
	userPrompt, messages, err := buildPredictionPrompts(ctx, stockCode, tradingStatusStr, predictionFocus, timeContextInstruction, data)
	if err != nil {
		return "", 0, "", "", 0, "", err
	}

	ctx = core.WithModelName(ctx, cfg.ModelName)
	ctx = core.WithGenerationName(ctx, "LLM-Predict")
	res, _, _, messages, err := p.generatePrediction(ctx, llm, stockCode, tradingStatusStr, data, messages, lf)
	if err != nil {
		return "", 0, "", "", 0, "", err
	}

	analysis, confidence, newsSummary, predictedChange, policyImpactScope, _ := parsePredictionResult(res)

	if lf != nil {
		traceID := langfuse.TraceIDFromContext(ctx)
		if traceID != "" {
			lf.UpdateTrace(ctx, traceID, userPrompt, analysis)
		}
	}

	if summaryText := formatEvidenceSummary(data.evidenceSummary); summaryText != "" {
		if newsSummary == "" || newsSummary == "详见分析" {
			newsSummary = summaryText
		} else {
			newsSummary = newsSummary + "\n\n" + summaryText
		}
	}

	return analysis, confidence, newsSummary, langfuse.TraceIDFromContext(ctx), predictedChange, policyImpactScope, nil
}

type predictionContext struct {
	stockData              string
	marketInfo             string
	analysisData           string
	sectorContext          string
	dtContext              string
	fractalAnalysisContext string
	knowledgeGraphEvidence string
	evidenceSummary        []string
	intradaySummary        string
	klineSummary           string
}

func (p *Provider) resolveModelConfig(modelName string) (core.ModelConfig, error) {
	if p.fileConfig == nil {
		return core.ModelConfig{}, fmt.Errorf("未找到配置且禁用了 fake 提供商")
	}

	var cfg core.ModelConfig
	found := false
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

	if !found {
		if modelName != "" {
			log.Printf("配置中未找到模型 %s，回退到当前提供商", modelName)
		}
		var ok bool
		cfg, ok = p.fileConfig.Models[string(p.fileConfig.CurrentProvider)]
		if ok {
			cfg.Provider = p.fileConfig.CurrentProvider
		} else {
			return core.ModelConfig{}, fmt.Errorf("未找到提供商且禁用了 fake 提供商")
		}
	}

	return cfg, nil
}

func startPredictionTrace(ctx context.Context, stockCode string) (context.Context, *langfuse.LangfuseManager) {
	if lf := langfuse.GetLangfuse(); lf != nil {
		traceID := lf.CreateTrace(ctx, "StockPrediction", map[string]interface{}{
			"stock_code": stockCode,
			"env":        "dev",
		})
		return langfuse.WithTraceID(ctx, traceID), lf
	}
	return ctx, nil
}

func (p *Provider) collectPredictionContext(ctx context.Context, stockCode string, days int32, lf *langfuse.LangfuseManager) predictionContext {
	var knowledgeGraphBundle *knowledge_graph.EvidenceBundle
	var knowledgeGraphEvidence string
	traceID := langfuse.TraceIDFromContext(ctx)
	log.Printf("collectPredictionContext start stock=%s trace=%s", stockCode, traceID)
	if lf != nil {
		start := time.Now()
		knowledgeGraphBundle, knowledgeGraphEvidence = fetchKnowledgeGraphEvidenceBundle(ctx, stockCode)
		if traceID != "" {
			lf.Span(ctx, traceID, nil, "Node:KnowledgeGraph", stockCode, knowledgeGraphEvidence, start, time.Now())
		}
	} else {
		knowledgeGraphBundle, knowledgeGraphEvidence = fetchKnowledgeGraphEvidenceBundle(ctx, stockCode)
	}
	if knowledgeGraphBundle != nil {
		log.Printf("KG bundle events=%d relations=%d entities=%d", len(knowledgeGraphBundle.Events), len(knowledgeGraphBundle.Relations), len(knowledgeGraphBundle.Entities))
	} else {
		log.Printf("KG bundle empty")
	}
	evidenceSummary := formatEvidenceSummary(buildEvidenceSummary(knowledgeGraphBundle))
	sectorFromKG := buildKGSectorContext(knowledgeGraphBundle)
	marketInfo := buildMarketInfoFromKG(knowledgeGraphBundle, evidenceSummary, knowledgeGraphEvidence, sectorFromKG)
	if marketInfo == "" || marketInfo == "暂无可用证据" {
		marketInfo = "暂无可用证据"
	}
	affiliation := buildStockAffiliationContext(ctx, stockCode)
	if strings.TrimSpace(affiliation) != "" {
		if marketInfo == "" || marketInfo == "暂无可用证据" {
			marketInfo = affiliation
		} else {
			marketInfo = strings.TrimSpace(affiliation) + "\n\n" + marketInfo
		}
	}
	log.Printf("KG evidence len=%d summary len=%d sector len=%d", len(knowledgeGraphEvidence), len(evidenceSummary), len(sectorFromKG))

	var stockData string
	analysisData := strings.TrimSpace(evidenceSummary)
	sectorContext := strings.TrimSpace(sectorFromKG)
	var dtContext string
	var fractalAnalysisContext string
	var intradaySummary string
	var klineSummary string
	stockTool := tool.NewStockPriceTool(rpc.StockClient)
	analysisTool := tool.NewStockAnalysisTool()
	sectorTool := tool.NewSectorTool(rpc.StockClient)
	dtTool := tool.NewDragonTigerTool()
	var errFetch error
	if lf != nil {
		start := time.Now()
		stockData, errFetch = stockTool.Call(ctx, stockCode)
		if traceID != "" {
			lf.Span(ctx, traceID, nil, "Node:StockPrice", stockCode, stockData, start, time.Now())
		}
	} else {
		stockData, errFetch = stockTool.Call(ctx, stockCode)
	}
	if errFetch != nil {
		log.Printf("预取股票数据失败: %v", errFetch)
		stockData = fmt.Sprintf("获取股票数据出错: %v", errFetch)
	}
	if analysisData == "" || analysisData == "暂无可用证据" {
		if lf != nil {
			start := time.Now()
			analysisData, _ = analysisTool.Call(ctx, stockCode)
			if traceID != "" {
				lf.Span(ctx, traceID, nil, "Node:Analysis", stockCode, analysisData, start, time.Now())
			}
		} else {
			analysisData, _ = analysisTool.Call(ctx, stockCode)
		}
	}
	log.Printf("analysis len=%d", len(analysisData))
	intradaySummary, intradayDate := buildIntradaySummary(ctx, stockCode, IsTradingTime())
	klineSummary = buildKlineSummary(ctx, stockCode)
	if intradaySummary != "" {
		writeIntradaySummaryToKG(ctx, stockCode, intradaySummary, intradayDate)
	}
	if sectorContext == "" {
		if lf != nil {
			start := time.Now()
			sectorContext, _ = sectorTool.Call(ctx, "industry")
			if traceID != "" {
				lf.Span(ctx, traceID, nil, "Node:Sector", "industry", sectorContext, start, time.Now())
			}
		} else {
			sectorContext, _ = sectorTool.Call(ctx, "industry")
		}
	}
	limitUpContext, _ := sectorTool.Call(ctx, "limit_up")
	if strings.TrimSpace(limitUpContext) != "" {
		if strings.TrimSpace(sectorContext) != "" {
			sectorContext = strings.TrimSpace(sectorContext) + "\n\n[涨停池摘要]\n" + strings.TrimSpace(limitUpContext)
		} else {
			sectorContext = "[涨停池摘要]\n" + strings.TrimSpace(limitUpContext)
		}
	}
	log.Printf("sector len=%d", len(sectorContext))
	if lf != nil {
		start := time.Now()
		dtContext, _ = dtTool.Call(ctx, "")
		if traceID != "" {
			lf.Span(ctx, traceID, nil, "Node:DragonTiger", "today", dtContext, start, time.Now())
		}
	} else {
		dtContext, _ = dtTool.Call(ctx, "")
	}
	log.Printf("dragonTiger len=%d", len(dtContext))
	startFractal := time.Now()
	fractalRes, _, _, err := p.predictWithFractal(ctx, stockCode, days)
	if err == nil {
		parts := strings.Split(fractalRes, "---METADATA---")
		if len(parts) > 0 {
			fractalAnalysisContext = strings.TrimSpace(parts[0])
		}
	} else {
		fractalAnalysisContext = "分形分析不可用。"
		log.Printf("分形分析失败: 股票=%s 天数=%d 错误=%v", stockCode, days, err)
	}
	if lf != nil {
		if traceID != "" {
			lf.Span(ctx, traceID, nil, "Node:Fractal", stockCode, fractalAnalysisContext, startFractal, time.Now())
		}
	}
	log.Printf("fractal len=%d", len(fractalAnalysisContext))

	return predictionContext{
		stockData:              stockData,
		marketInfo:             marketInfo,
		analysisData:           analysisData,
		sectorContext:          sectorContext,
		dtContext:              dtContext,
		fractalAnalysisContext: fractalAnalysisContext,
		knowledgeGraphEvidence: knowledgeGraphEvidence,
		evidenceSummary:        buildEvidenceSummary(knowledgeGraphBundle),
		intradaySummary:        intradaySummary,
		klineSummary:           klineSummary,
	}
}

func buildIntradaySummary(ctx context.Context, stockCode string, isTrading bool) (string, string) {
	client := eastmoney.NewClient()
	klines, err := client.GetKlineHistoryWithKlt(ctx, stockCode, 800, 1)
	if err != nil || len(klines) == 0 {
		return "", ""
	}
	grouped := make(map[string][]*eastmoney.KlineItem)
	order := make([]string, 0, 3)
	for _, k := range klines {
		if k == nil || len(k.Date) < 10 {
			continue
		}
		date := k.Date[:10]
		if _, ok := grouped[date]; !ok {
			order = append(order, date)
		}
		grouped[date] = append(grouped[date], k)
	}
	if len(order) == 0 {
		return "", ""
	}
	if len(order) > 3 {
		order = order[len(order)-3:]
	}
	var daySummaries []string
	var upDays int
	var downDays int
	var totalVolume int64
	var totalChange float64
	var latestDate string
	var latestSummary string
	for _, date := range order {
		items := grouped[date]
		if len(items) == 0 {
			continue
		}
		open := items[0].Open
		close := items[len(items)-1].Close
		high := items[0].High
		low := items[0].Low
		var volume int64
		for _, it := range items {
			if it.High > high {
				high = it.High
			}
			if it.Low < low {
				low = it.Low
			}
			volume += it.Volume
		}
		changePct := 0.0
		if open > 0 {
			changePct = (close - open) / open * 100
		}
		totalChange += changePct
		totalVolume += volume
		if changePct > 0.2 {
			upDays++
		} else if changePct < -0.2 {
			downDays++
		}
		status := "震荡"
		if changePct > 0.5 {
			status = "偏强"
		} else if changePct < -0.5 {
			status = "偏弱"
		}
		summary := fmt.Sprintf("%s: 开盘%.2f 收盘%.2f 高%.2f 低%.2f 涨跌%.2f%% 量能%.0f万股 走势%s", date, open, close, high, low, changePct, float64(volume)/10000, status)
		daySummaries = append(daySummaries, summary)
		latestDate = date
		latestSummary = summary
	}
	if len(daySummaries) == 0 {
		return "", ""
	}
	overallTrend := "震荡"
	if upDays >= 2 {
		overallTrend = "偏强"
	} else if downDays >= 2 {
		overallTrend = "偏弱"
	}
	avgChange := totalChange / float64(len(daySummaries))
	avgVolume := float64(totalVolume) / float64(len(daySummaries)) / 10000
	latestStatus := latestTradingStatus(isTrading, latestDate)
	overall := fmt.Sprintf("近3日整体分时表现: 趋势%s，平均涨跌%.2f%%，日均量能%.0f万股", overallTrend, avgChange, avgVolume)
	return fmt.Sprintf("%s\n最新交易日(%s, %s)表现: %s\n明细:\n- %s", overall, latestDate, latestStatus, latestSummary, strings.Join(daySummaries, "\n- ")), latestDate
}

func latestTradingStatus(isTrading bool, latestDate string) string {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	now := time.Now().In(loc)
	today := now.Format("2006-01-02")
	if latestDate != today {
		return "已收盘"
	}
	if isTrading {
		return "盘中"
	}
	return "已收盘"
}

func buildKlineSummary(ctx context.Context, stockCode string) string {
	client := eastmoney.NewClient()
	klines, err := client.GetKlineHistory(ctx, stockCode, 5)
	if err != nil || len(klines) == 0 {
		return ""
	}
	if len(klines) > 5 {
		klines = klines[len(klines)-5:]
	}
	first := klines[0]
	last := klines[len(klines)-1]
	trend := "震荡"
	if last.Close > first.Open {
		trend = "偏多"
	} else if last.Close < first.Open {
		trend = "偏空"
	}
	upStreak := 0
	downStreak := 0
	for _, k := range klines {
		if k.Close >= k.Open {
			upStreak++
			downStreak = 0
		} else {
			downStreak++
			upStreak = 0
		}
	}
	pattern := "无明显连续形态"
	if upStreak >= 3 {
		pattern = fmt.Sprintf("%d 连阳", upStreak)
	} else if downStreak >= 3 {
		pattern = fmt.Sprintf("%d 连阴", downStreak)
	}
	shadow := describeShadow(last)
	volumeTrend := describeVolumeTrend(klines)
	return fmt.Sprintf("近5日趋势%s，形态%s，末日蜡烛%s，量能%s", trend, pattern, shadow, volumeTrend)
}

func describeShadow(k *eastmoney.KlineItem) string {
	if k == nil {
		return "无数据"
	}
	body := math.Abs(k.Close - k.Open)
	upper := k.High - math.Max(k.Open, k.Close)
	lower := math.Min(k.Open, k.Close) - k.Low
	if body == 0 {
		body = 0.01
	}
	if upper > body*1.5 && lower > body*1.5 {
		return "长上影+长下影"
	}
	if upper > body*1.5 {
		return "长上影"
	}
	if lower > body*1.5 {
		return "长下影"
	}
	return "影线较短"
}

func describeVolumeTrend(klines []*eastmoney.KlineItem) string {
	if len(klines) < 2 {
		return "量能平稳"
	}
	first := klines[0].Volume
	last := klines[len(klines)-1].Volume
	if first == 0 {
		return "量能平稳"
	}
	ratio := float64(last) / float64(first)
	if ratio > 1.2 {
		return "放量"
	}
	if ratio < 0.8 {
		return "缩量"
	}
	return "量能平稳"
}

func writeIntradaySummaryToKG(ctx context.Context, stockCode string, summary string, latestDate string) {
	if summary == "" || IsTradingTime() {
		return
	}
	if rpc.KnowledgeGraphClient == nil {
		return
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	now := time.Now().In(loc)
	date := latestDate
	if date == "" {
		date = now.Format("2006-01-02")
	}
	id := fmt.Sprintf("intraday_summary:%s:%s", stockCode, date)
	impactDirection := int32(0)
	if strings.Contains(summary, "偏强") {
		impactDirection = 1
	} else if strings.Contains(summary, "偏弱") {
		impactDirection = -1
	}
	impactStrength := int32(50)
	confidence := int32(70)
	event := &knowledge_graph.Event{
		Id:              id,
		Type:            knowledge_graph.EventType_REPORT,
		Entities:        []*knowledge_graph.Entity{{Id: stockCode, Type: knowledge_graph.EntityType_STOCK, Name: stockCode}},
		ImpactDirection: impactDirection,
		ImpactStrength:  impactStrength,
		Confidence:      confidence,
		Timestamp:       now.Unix(),
		Source:          "intraday_summary",
		DedupeKey:       id,
	}
	_, _ = rpc.KnowledgeGraphClient.UpsertEvents(ctx, &knowledge_graph.UpsertEventsRequest{Events: []*knowledge_graph.Event{event}})
}

func buildKGSectorContext(bundle *knowledge_graph.EvidenceBundle) string {
	if bundle == nil || len(bundle.Entities) == 0 {
		return ""
	}
	names := make(map[string]struct{})
	for _, ent := range bundle.Entities {
		if ent == nil || ent.Type != knowledge_graph.EntityType_SECTOR {
			continue
		}
		name := strings.TrimSpace(ent.Name)
		if name == "" && ent.Attributes != nil {
			name = strings.TrimSpace(ent.Attributes["name"])
		}
		if name == "" {
			continue
		}
		names[name] = struct{}{}
	}
	if len(names) == 0 {
		return ""
	}
	list := make([]string, 0, len(names))
	for name := range names {
		list = append(list, name)
	}
	sort.Strings(list)
	return "知识图谱板块: " + strings.Join(list, "、")
}

func buildMarketInfoFromKG(bundle *knowledge_graph.EvidenceBundle, summaryText string, rawEvidence string, sectorText string) string {
	eventsText := formatKGEvents(bundle, 6)
	relationsText := formatKGRelations(bundle, 6)
	cleanSummary := strings.TrimSpace(summaryText)
	if cleanSummary == "暂无可用证据" || strings.Contains(cleanSummary, "暂无可用证据") {
		cleanSummary = ""
	}
	rawEvidence = strings.TrimSpace(rawEvidence)
	if rawEvidence == "暂无可用证据" {
		rawEvidence = ""
	}
	var builder strings.Builder
	if eventsText != "" {
		builder.WriteString("市场事件:\n")
		builder.WriteString(eventsText)
		builder.WriteString("\n")
	}
	if relationsText != "" {
		builder.WriteString("关系线索:\n")
		builder.WriteString(relationsText)
		builder.WriteString("\n")
	}
	if sectorText != "" {
		builder.WriteString("板块关联:\n")
		builder.WriteString(strings.TrimSpace(sectorText))
		builder.WriteString("\n")
	}
	if cleanSummary != "" {
		builder.WriteString(cleanSummary)
		builder.WriteString("\n")
	}
	if builder.Len() == 0 {
		if rawEvidence != "" {
			return rawEvidence
		}
		if cleanSummary != "" {
			return cleanSummary
		}
		return ""
	}
	return strings.TrimSpace(builder.String())
}

func formatKGEvents(bundle *knowledge_graph.EvidenceBundle, limit int) string {
	if bundle == nil || len(bundle.Events) == 0 || limit <= 0 {
		return ""
	}
	events := make([]*knowledge_graph.Event, 0, len(bundle.Events))
	for _, ev := range bundle.Events {
		if ev != nil {
			events = append(events, ev)
		}
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp > events[j].Timestamp
	})
	titleByID, summaryByID := buildEventEvidenceIndex(bundle)
	var lines []string
	for _, ev := range events {
		day := time.Unix(ev.Timestamp, 0).Format("01-02")
		title := strings.TrimSpace(titleByID[ev.Id])
		summary := strings.TrimSpace(summaryByID[ev.Id])
		if title != "" {
			title = trimText(compactText(title), 36)
		}
		if summary != "" {
			summary = trimText(compactText(summary), 80)
		}
		content := strings.TrimSpace(strings.TrimSpace(title + " " + summary))
		if content == "" {
			content = fmt.Sprintf("影响:%d 强度:%d 可信:%d", ev.ImpactDirection, ev.ImpactStrength, ev.Confidence)
		}
		line := fmt.Sprintf("- %s %s %s 来源:%s", day, ev.Type.String(), content, ev.Source)
		lines = append(lines, line)
		if len(lines) >= limit {
			break
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func formatKGRelations(bundle *knowledge_graph.EvidenceBundle, limit int) string {
	if bundle == nil || len(bundle.Relations) == 0 || limit <= 0 {
		return ""
	}
	relations := make([]*knowledge_graph.Relation, 0, len(bundle.Relations))
	for _, rel := range bundle.Relations {
		if rel != nil {
			relations = append(relations, rel)
		}
	}
	sort.Slice(relations, func(i, j int) bool {
		return relations[i].Strength > relations[j].Strength
	})
	var lines []string
	for _, rel := range relations {
		source := formatEntityName(rel.Source)
		target := formatEntityName(rel.Target)
		if source == "" {
			source = "-"
		}
		if target == "" {
			target = "-"
		}
		summary := ""
		if rel.Evidence != nil {
			summary = strings.TrimSpace(rel.Evidence.Summary)
			if summary != "" {
				summary = trimText(compactText(summary), 80)
			}
		}
		line := fmt.Sprintf("- %s -> %s | %s | 强度:%.2f", source, target, rel.Type.String(), rel.Strength)
		if summary != "" {
			line += fmt.Sprintf(" | 证据:%s", summary)
		}
		lines = append(lines, line)
		if len(lines) >= limit {
			break
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func formatEntityName(ent *knowledge_graph.Entity) string {
	if ent == nil {
		return ""
	}
	name := strings.TrimSpace(ent.Name)
	if name == "" && ent.Attributes != nil {
		name = strings.TrimSpace(ent.Attributes["name"])
	}
	if name != "" {
		return name
	}
	return strings.TrimSpace(ent.Id)
}

func boardNameOf(code string) string {
	if strings.HasPrefix(code, "sh688") {
		return "科创板"
	}
	if strings.HasPrefix(code, "sz300") {
		return "创业板"
	}
	if strings.HasPrefix(code, "bj") {
		return "北证"
	}
	if strings.HasPrefix(code, "sh") {
		return "上证"
	}
	if strings.HasPrefix(code, "sz") {
		return "深证"
	}
	return ""
}

func buildStockAffiliationContext(ctx context.Context, stockCode string) string {
	client := eastmoney.NewClient()
	ind, err := client.GetIndustryIndex(ctx, stockCode)
	industry := ""
	region := ""
	concepts := ""
	if err == nil && ind != nil {
		industry = strings.TrimSpace(ind.IndustryName)
		region = strings.TrimSpace(ind.RegionName)
		concepts = strings.TrimSpace(ind.ConceptNames)
	}
	board := boardNameOf(strings.TrimSpace(stockCode))
	parts := make([]string, 0, 4)
	if industry != "" {
		parts = append(parts, fmt.Sprintf("行业: %s", industry))
	}
	if concepts != "" {
		parts = append(parts, fmt.Sprintf("概念: %s", concepts))
	}
	if region != "" {
		parts = append(parts, fmt.Sprintf("地区: %s", region))
	}
	if board != "" {
		parts = append(parts, fmt.Sprintf("交易板块: %s", board))
	}
	if len(parts) == 0 {
		return ""
	}
	return "个股归属:\n" + strings.Join(parts, "；")
}

func buildTradingContext() (string, string, string) {
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
	return tradingStatusStr, predictionFocus, timeContextInstruction
}

func buildPredictionPrompts(ctx context.Context, stockCode, tradingStatusStr, predictionFocus, timeContextInstruction string, data predictionContext) (string, []llms.MessageContent, error) {
	systemPromptTemplate := prompt.GetManager().GetPrompt(ctx, prompt.StockPredictionSystem)
	systemPrompt := fmt.Sprintf(systemPromptTemplate, timeContextInstruction, time.Now().Format("2006-01-02 15:04:05"), predictionFocus)
	userPromptTemplate := prompt.GetManager().GetPrompt(ctx, prompt.StockPredictionUser)
	if userPromptTemplate == "" {
		return "", nil, fmt.Errorf("获取预测用户提示词失败")
	}
	userPrompt := buildUserPrompt(userPromptTemplate, stockCode, tradingStatusStr, data.stockData, data.analysisData, data.marketInfo, data.sectorContext, data.dtContext, data.fractalAnalysisContext, data.intradaySummary, data.klineSummary, data.knowledgeGraphEvidence)
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
	return userPrompt, messages, nil
}

func (p *Provider) generatePrediction(ctx context.Context, llm llms.Model, stockCode, tradingStatusStr string, data predictionContext, messages []llms.MessageContent, lf *langfuse.LangfuseManager) (string, time.Time, time.Time, []llms.MessageContent, error) {
	var res string
	var genStart time.Time
	var genEnd time.Time

	enableMultiAgent := false
	cfg := config.Get()
	if cfg != nil && cfg.Runtime != nil && cfg.Runtime.MultiAgentEnabled {
		enableMultiAgent = true
	}
	if enableMultiAgent {
		orchestratedRes, orchestratedPrompt, oStart, oEnd, err := p.runMultiAgentPrediction(ctx, llm, stockCode, tradingStatusStr, data.stockData, data.analysisData, data.marketInfo, data.sectorContext, data.dtContext, data.fractalAnalysisContext, data.intradaySummary, data.klineSummary, data.knowledgeGraphEvidence, lf)
		if err == nil {
			res = orchestratedRes
			genStart = oStart
			genEnd = oEnd
			messages = []llms.MessageContent{
				{
					Role:  llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{llms.TextContent{Text: orchestratedPrompt}},
				},
			}
		} else {
			log.Printf("多智能体协作失败，回退单模型: %v", err)
		}
	}

	if res == "" {
		genStart = time.Now()
		resp, err := core.GenerateContentWithRetry(ctx, llm, messages)
		genEnd = time.Now()
		if err != nil {
			return "", time.Time{}, time.Time{}, messages, fmt.Errorf("llm 生成失败: %w", err)
		}
		if len(resp.Choices) == 0 {
			return "", time.Time{}, time.Time{}, messages, fmt.Errorf("llm 响应为空")
		}
		res = resp.Choices[0].Content
	}

	return res, genStart, genEnd, messages, nil
}

func parsePredictionResult(res string) (string, float64, string, float64, string, map[string]interface{}) {
	analysis := res
	confidence := 0.5
	newsSummary := "详见分析"
	policyImpactScope := "specific"
	predictedChange := 0.0
	parts := strings.Split(res, "---METADATA---")
	var metadataMap map[string]interface{}
	if len(parts) > 1 {
		analysis = strings.TrimSpace(parts[0])
		metadataJSON := strings.TrimSpace(parts[1])
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
	return analysis, confidence, newsSummary, predictedChange, policyImpactScope, metadataMap
}

func buildLangfuseInput(messages []llms.MessageContent) []map[string]interface{} {
	lfInput := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		var contentStr string
		for _, part := range msg.Parts {
			if textPart, ok := part.(llms.TextContent); ok {
				contentStr += textPart.Text
			} else if imgPart, ok := part.(llms.ImageURLContent); ok {
				contentStr += fmt.Sprintf("[Image: %s]", imgPart.URL)
			}
		}
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
	return lfInput
}

func fetchKnowledgeGraphEvidenceBundle(ctx context.Context, stockCode string) (*knowledge_graph.EvidenceBundle, string) {
	if rpc.KnowledgeGraphClient == nil {
		return nil, "暂无可用证据"
	}

	endTime := time.Now().Unix()
	startTime := time.Now().AddDate(0, 0, -180).Unix()
	req := &knowledge_graph.GetEvidenceBundleRequest{
		StockCode:     stockCode,
		StartTime:     startTime,
		EndTime:       endTime,
		MinConfidence: 0,
		MaxItems:      50,
	}
	resp, err := rpc.KnowledgeGraphClient.GetEvidenceBundle(ctx, req)
	if err != nil || resp == nil || resp.Bundle == nil {
		return nil, "暂无可用证据"
	}

	return resp.Bundle, formatEvidenceBundle(resp.Bundle)
}

func compactText(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

func trimText(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + "..."
}

func buildEventEvidenceIndex(bundle *knowledge_graph.EvidenceBundle) (map[string]string, map[string]string) {
	titleByID := make(map[string]string)
	summaryByID := make(map[string]string)
	if bundle == nil {
		return titleByID, summaryByID
	}
	for _, ent := range bundle.Entities {
		if ent == nil || ent.Type != knowledge_graph.EntityType_EVENT || ent.Id == "" {
			continue
		}
		if ent.Name != "" {
			titleByID[ent.Id] = ent.Name
		}
	}
	for _, rel := range bundle.Relations {
		if rel == nil {
			continue
		}
		if rel.Source != nil && rel.Source.Type == knowledge_graph.EntityType_EVENT && rel.Source.Id != "" && rel.Source.Name != "" {
			if titleByID[rel.Source.Id] == "" {
				titleByID[rel.Source.Id] = rel.Source.Name
			}
		}
		if rel.Target != nil && rel.Target.Type == knowledge_graph.EntityType_EVENT && rel.Target.Id != "" && rel.Target.Name != "" {
			if titleByID[rel.Target.Id] == "" {
				titleByID[rel.Target.Id] = rel.Target.Name
			}
		}
		if rel.Evidence == nil {
			continue
		}
		summary := strings.TrimSpace(rel.Evidence.Summary)
		if summary == "" {
			continue
		}
		if rel.Source != nil && rel.Source.Type == knowledge_graph.EntityType_EVENT && rel.Source.Id != "" {
			if summaryByID[rel.Source.Id] == "" {
				summaryByID[rel.Source.Id] = summary
			}
		}
		if rel.Target != nil && rel.Target.Type == knowledge_graph.EntityType_EVENT && rel.Target.Id != "" {
			if summaryByID[rel.Target.Id] == "" {
				summaryByID[rel.Target.Id] = summary
			}
		}
	}
	return titleByID, summaryByID
}

func formatEvidenceBundle(bundle *knowledge_graph.EvidenceBundle) string {
	var builder strings.Builder
	titleByID, summaryByID := buildEventEvidenceIndex(bundle)
	if len(bundle.Events) > 0 {
		builder.WriteString("事件证据:\n")
		for _, ev := range bundle.Events {
			if ev == nil {
				continue
			}
			day := time.Unix(ev.Timestamp, 0).Format("2006-01-02")
			line := fmt.Sprintf("- %s | %s | 影响:%d 强度:%d 可信:%d 来源:%s",
				day, ev.Type.String(), ev.ImpactDirection, ev.ImpactStrength, ev.Confidence, ev.Source)
			title := strings.TrimSpace(titleByID[ev.Id])
			if title != "" {
				title = trimText(compactText(title), 80)
				line += fmt.Sprintf(" 标题:%s", title)
			}
			summary := strings.TrimSpace(summaryByID[ev.Id])
			if summary != "" {
				summary = trimText(compactText(summary), 160)
				if summary != title {
					line += fmt.Sprintf(" 内容:%s", summary)
				}
			}
			builder.WriteString(line + "\n")
		}
	}
	if len(bundle.Relations) > 0 {
		if builder.Len() > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString("关系证据:\n")
		for _, rel := range bundle.Relations {
			if rel == nil {
				continue
			}
			sourceID := ""
			targetID := ""
			if rel.Source != nil {
				sourceID = rel.Source.Id
			}
			if rel.Target != nil {
				targetID = rel.Target.Id
			}
			summary := ""
			if rel.Evidence != nil {
				summary = rel.Evidence.Summary
			}
			if summary != "" {
				builder.WriteString(fmt.Sprintf("- %s -> %s | %s | 强度:%.2f | 证据:%s\n", sourceID, targetID, rel.Type.String(), rel.Strength, summary))
			} else {
				builder.WriteString(fmt.Sprintf("- %s -> %s | %s | 强度:%.2f\n", sourceID, targetID, rel.Type.String(), rel.Strength))
			}
		}
	}
	if builder.Len() == 0 {
		return "暂无可用证据"
	}
	return strings.TrimSpace(builder.String())
}

func buildEvidenceSummary(bundle *knowledge_graph.EvidenceBundle) []string {
	if bundle == nil {
		return []string{"暂无可用证据"}
	}
	titleByID, summaryByID := buildEventEvidenceIndex(bundle)
	items := make([]string, 0, 5)
	for _, ev := range bundle.Events {
		if ev == nil {
			continue
		}
		day := time.Unix(ev.Timestamp, 0).Format("01-02")
		title := strings.TrimSpace(titleByID[ev.Id])
		summary := strings.TrimSpace(summaryByID[ev.Id])
		if summary != "" {
			summary = trimText(compactText(summary), 80)
		}
		if title != "" {
			title = trimText(compactText(title), 40)
		}
		if summary != "" && summary != title {
			items = append(items, fmt.Sprintf("%s %s %s", day, ev.Type.String(), summary))
		} else if title != "" {
			items = append(items, fmt.Sprintf("%s %s %s", day, ev.Type.String(), title))
		} else {
			items = append(items, fmt.Sprintf("%s %s 影响:%d 强度:%d 可信:%d 来源:%s", day, ev.Type.String(), ev.ImpactDirection, ev.ImpactStrength, ev.Confidence, ev.Source))
		}
		if len(items) >= 5 {
			return items
		}
	}
	for _, rel := range bundle.Relations {
		if rel == nil {
			continue
		}
		sourceID := ""
		targetID := ""
		if rel.Source != nil {
			sourceID = rel.Source.Id
		}
		if rel.Target != nil {
			targetID = rel.Target.Id
		}
		summary := ""
		if rel.Evidence != nil {
			summary = rel.Evidence.Summary
		}
		if summary != "" {
			items = append(items, fmt.Sprintf("%s->%s %s 强度:%.2f 证据:%s", sourceID, targetID, rel.Type.String(), rel.Strength, summary))
		} else {
			items = append(items, fmt.Sprintf("%s->%s %s 强度:%.2f", sourceID, targetID, rel.Type.String(), rel.Strength))
		}
		if len(items) >= 5 {
			return items
		}
	}
	return items
}

func formatEvidenceSummary(items []string) string {
	if len(items) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("证据摘要:\n")
	for _, item := range items {
		builder.WriteString("- ")
		builder.WriteString(item)
		builder.WriteString("\n")
	}
	return strings.TrimSpace(builder.String())
}

func buildUserPrompt(template, stockCode, tradingStatus, stockData, analysisData, marketInfo, sectorContext, dtContext, fractalContext, intradaySummary, klineSummary, evidence string) string {
	placeholderCount := strings.Count(template, "%s")
	args := []interface{}{
		stockCode,
		time.Now().Format("2006-01-02 15:04:05"),
		tradingStatus,
		stockData,
		analysisData,
		marketInfo,
		sectorContext,
		dtContext,
		fractalContext,
		intradaySummary,
		klineSummary,
	}
	if placeholderCount >= 12 {
		args = append(args, evidence)
		return fmt.Sprintf(template, args...)
	}
	promptText := fmt.Sprintf(template, args...)
	return promptText + "\n\n[知识图谱证据]:\n" + evidence
}

func (p *Provider) predictWithFractal(ctx context.Context, stockCode string, days int32) (string, float64, string, error) {
	// 1. 获取历史 K 线 (例如 500 天)
	req := &stock.GetHistoricalKlineRequest{
		StockCode: stockCode,
		Days:      500,
	}
	var klines []*stock.Kline
	var err error
	if rpc.StockClient != nil {
		resp, rpcErr := rpc.StockClient.GetHistoricalKline(ctx, req)
		if rpcErr == nil && resp != nil && len(resp.Klines) > 0 {
			klines = resp.Klines
		} else {
			err = rpcErr
		}
	} else {
		err = fmt.Errorf("Stock Service 未初始化")
	}
	if len(klines) == 0 {
		em := eastmoney.NewClient()
		items, emErr := em.GetKlineHistory(ctx, stockCode, int(req.Days))
		if (emErr != nil || len(items) == 0) && int(req.Days) > 300 {
			items, emErr = em.GetKlineHistory(ctx, stockCode, 300)
		}
		if (emErr != nil || len(items) == 0) && int(req.Days) > 200 {
			items, emErr = em.GetKlineHistory(ctx, stockCode, 200)
		}
		if emErr != nil || len(items) == 0 {
			return "", 0, "", fmt.Errorf("获取K线失败: %v", func() error {
				if err != nil {
					return fmt.Errorf("rpc: %v, eastmoney: %v", err, emErr)
				}
				return emErr
			}())
		}
		klines = make([]*stock.Kline, 0, len(items))
		for _, k := range items {
			if k == nil {
				continue
			}
			klines = append(klines, &stock.Kline{
				Date:   k.Date,
				Open:   k.Open,
				Close:  k.Close,
				High:   k.High,
				Low:    k.Low,
				Volume: k.Volume,
			})
		}
	}
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
