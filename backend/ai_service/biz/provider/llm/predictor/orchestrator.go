package predictor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"stock_assistant/backend/ai_service/biz/provider/llm/core"
	"stock_assistant/backend/ai_service/biz/provider/prompt"
	"stock_assistant/backend/common/langfuse"

	"github.com/tmc/langchaingo/llms"
)

type AgentOutput struct {
	Direction  string   `json:"direction"`
	Confidence float64  `json:"confidence"`
	Summary    string   `json:"summary"`
	Evidence   []string `json:"evidence"`
}

type AgentRunResult struct {
	Role   string      `json:"role"`
	Output AgentOutput `json:"output"`
	Error  string      `json:"error"`
	Raw    string      `json:"raw"`
}

type CoordinatorInput struct {
	StockCode     string           `json:"stock_code"`
	Time          string           `json:"time"`
	TradingStatus string           `json:"trading_status"`
	Results       []AgentRunResult `json:"results"`
}

func (p *Provider) runMultiAgentPrediction(ctx context.Context, llm llms.Model, stockCode string, tradingStatusStr string, stockData string, analysisData string, marketInfo string, sectorContext string, dtContext string, fractalAnalysisContext string, intradaySummary string, klineSummary string, knowledgeGraphEvidence string, lf *langfuse.LangfuseManager) (string, string, time.Time, time.Time, error) {
	rootCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	traceID := langfuse.TraceIDFromContext(ctx)
	timeStr := time.Now().Format("2006-01-02 15:04:05")
	fundamentalContext := fmt.Sprintf("财务与估值信息:\n%s\n\n公司分析:\n%s\n\n行业对比:\n%s", stockData, analysisData, sectorContext)
	eventContext := fmt.Sprintf("事件证据:\n%s\n\n市场新闻:\n%s", knowledgeGraphEvidence, marketInfo)
	technicalRules := strings.TrimSpace(prompt.GetManager().GetPrompt(rootCtx, prompt.TechnicalRules))
	technicalContext := fmt.Sprintf("技术面数据:\n%s\n\n%s\n\n分形分析:\n%s\n\n近3日分时表现摘要:\n%s\n\n近5日K线/量价摘要:\n%s", stockData, analysisData, fractalAnalysisContext, intradaySummary, klineSummary)
	if technicalRules != "" {
		technicalContext = fmt.Sprintf("%s\n\n技术面规则库:\n%s", technicalContext, technicalRules)
	}
	riskContext := fmt.Sprintf("风险相关信息:\n市场信息:\n%s\n\n龙虎榜:\n%s\n\n事件证据:\n%s", marketInfo, dtContext, knowledgeGraphEvidence)

	type agentCall struct {
		role      string
		promptKey string
		context   string
	}

	calls := []agentCall{
		{role: "基本面智能体", promptKey: prompt.FundamentalAgent, context: fundamentalContext},
		{role: "事件智能体", promptKey: prompt.EventAgent, context: eventContext},
		{role: "技术面智能体", promptKey: prompt.TechnicalAgent, context: technicalContext},
		{role: "风险智能体", promptKey: prompt.RiskAgent, context: riskContext},
	}

	results := make([]AgentRunResult, 0, len(calls))
	for _, call := range calls {
		agentCtx, agentCancel := context.WithTimeout(rootCtx, 3*time.Minute)
		template := prompt.GetManager().GetPrompt(agentCtx, call.promptKey)
		if template == "" {
			results = append(results, AgentRunResult{Role: call.role, Error: "提示词缺失"})
			agentCancel()
			continue
		}
		promptStr := fmt.Sprintf(template, stockCode, timeStr, call.context)
		messages := []llms.MessageContent{
			{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextContent{Text: promptStr}},
			},
		}
		agentCtx = core.WithGenerationName(agentCtx, "LLM-"+strings.ReplaceAll(call.role, "智能体", "Agent"))
		start := time.Now()
		resp, err := core.GenerateContentWithRetry(agentCtx, llm, messages)
		end := time.Now()
		raw := ""
		if err == nil && resp != nil && len(resp.Choices) > 0 {
			raw = resp.Choices[0].Content
		}
		if lf != nil && traceID != "" {
			lf.Span(agentCtx, traceID, nil, "Node:"+strings.ReplaceAll(call.role, "智能体", "Agent"), promptStr, raw, start, end)
		}
		if err != nil {
			results = append(results, AgentRunResult{Role: call.role, Error: err.Error(), Raw: raw})
			agentCancel()
			continue
		}
		output, parseErr := parseAgentOutput(raw)
		if parseErr != nil {
			results = append(results, AgentRunResult{Role: call.role, Error: parseErr.Error(), Raw: raw})
			agentCancel()
			continue
		}
		results = append(results, AgentRunResult{Role: call.role, Output: output, Raw: raw})
		agentCancel()
	}

	coordTemplate := prompt.GetManager().GetPrompt(rootCtx, prompt.CoordinatorAgent)
	if coordTemplate == "" {
		return "", "", time.Time{}, time.Time{}, fmt.Errorf("协调智能体提示词缺失")
	}
	input := CoordinatorInput{
		StockCode:     stockCode,
		Time:          timeStr,
		TradingStatus: tradingStatusStr,
		Results:       results,
	}
	inputJSON, _ := json.Marshal(input)
	coordPrompt := fmt.Sprintf(coordTemplate, stockCode, timeStr, tradingStatusStr, string(inputJSON))
	coordMessages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: coordPrompt}},
		},
	}
	rootCtx = core.WithGenerationName(rootCtx, "LLM-Coordinator")
	genStart := time.Now()
	coordResp, err := core.GenerateContentWithRetry(rootCtx, llm, coordMessages)
	genEnd := time.Now()
	if err != nil {
		return "", "", genStart, genEnd, err
	}
	if coordResp == nil || len(coordResp.Choices) == 0 {
		return "", "", genStart, genEnd, fmt.Errorf("协调智能体响应为空")
	}
	coordOutput := coordResp.Choices[0].Content
	if lf != nil && traceID != "" {
		lf.Span(rootCtx, traceID, nil, "Node:Coordinator", coordPrompt, coordOutput, genStart, genEnd)
	}
	return coordOutput, coordPrompt, genStart, genEnd, nil
}

func parseAgentOutput(raw string) (AgentOutput, error) {
	content := strings.TrimSpace(raw)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)
	var output AgentOutput
	if err := json.Unmarshal([]byte(content), &output); err != nil {
		return AgentOutput{}, err
	}
	return output, nil
}
