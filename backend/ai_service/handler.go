package main

import (
	"context"
	"fmt"
	"log"
	"stock_assistant/backend/ai_service/biz/provider/llm/analyst"
	"stock_assistant/backend/ai_service/biz/provider/llm/core"
	"stock_assistant/backend/ai_service/biz/provider/llm/predictor"
	"stock_assistant/backend/ai_service/biz/provider/llm/vision"
	"stock_assistant/backend/ai_service/biz/rpc"
	ai "stock_assistant/backend/ai_service/kitex_gen/ai"
	"stock_assistant/backend/ai_service/kitex_gen/stock"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AIServiceImpl implements the last service interface defined in the IDL.
type AIServiceImpl struct {
	predictor *predictor.Provider
	analyst   *analyst.Provider
	vision    *vision.Provider
}

func NewAIServiceImpl(llmConfig *core.FileConfig) *AIServiceImpl {
	pred := predictor.NewProvider(llmConfig)
	ana := analyst.NewProvider(llmConfig)
	vis := vision.NewProvider(llmConfig)

	return &AIServiceImpl{
		predictor: pred,
		analyst:   ana,
		vision:    vis,
	}
}

// GetPrediction implements the AIServiceImpl interface.
func (s *AIServiceImpl) GetPrediction(ctx context.Context, req *ai.GetPredictionRequest) (resp *ai.GetPredictionResponse, err error) {
	log.Printf("收到预测请求: 代码=%s, 模型=%s", req.Code, req.Model)
	analysis, confidence, newsSummary, traceID, predictedChange, policyImpactScope, err := s.predictor.Predict(ctx, req.Code, req.Days, req.Model)
	if err != nil {
		log.Printf("预测失败: %v", err)
		return nil, err
	}

	// 保存预测结果
	trend := "中性"
	if strings.Contains(analysis, "看涨") || strings.Contains(analysis, "Up") {
		trend = "看涨"
	} else if strings.Contains(analysis, "看跌") || strings.Contains(analysis, "Down") {
		trend = "看跌"
	}

	saveReq := &stock.SavePredictionRequest{
		Record: &stock.PredictionRecord{
			Id:                uuid.New().String(),
			StockCode:         req.Code,
			PredictionDate:    time.Now().Format("2006-01-02 15:04:05"),
			Content:           analysis,
			Confidence:        confidence,
			Trend:             trend,
			TraceId:           traceID,
			PolicyImpactScope: policyImpactScope,
			PredictedChange:   predictedChange,
		},
	}

	_, saveErr := rpc.StockClient.SavePrediction(ctx, saveReq)
	if saveErr != nil {
		log.Printf("保存预测结果失败: %v", saveErr)
	}

	return &ai.GetPredictionResponse{
		Result_: &ai.PredictionResult_{
			Code:         req.Code,
			Confidence:   confidence,
			Analysis:     analysis,
			NewsSummary_: newsSummary,
			TraceId:      traceID,
		},
	}, nil
}

// ImageRecognition implements the AIServiceImpl interface.
func (s *AIServiceImpl) ImageRecognition(ctx context.Context, req *ai.ImageRecognitionRequest) (resp *ai.ImageRecognitionResponse, err error) {
	log.Printf("收到图像识别请求: 模型=%s, 图片大小=%d", req.Model, len(req.ImageData))

	stocks, err := s.vision.RecognizeImage(ctx, req.ImageData, req.Model)
	if err != nil {
		log.Printf("图像识别失败: %v", err)
		return nil, err
	}

	return &ai.ImageRecognitionResponse{
		Stocks: stocks,
	}, nil
}

// MarketReview implements the AIServiceImpl interface.
func (s *AIServiceImpl) MarketReview(ctx context.Context, req *ai.MarketReviewRequest) (resp *ai.MarketReviewResponse, err error) {
	log.Printf("收到市场复盘请求: 日期=%s", req.Date)

	// 1. 获取板块数据
	sectorReq := &stock.GetMarketSectorsRequest{
		Type:  "concept",
		Limit: 20,
	}
	sectorResp, err := rpc.StockClient.GetMarketSectors(ctx, sectorReq)
	if err != nil {
		log.Printf("获取市场板块失败: %v", err)
		return nil, err
	}

	// 2. 获取涨停数据
	limitUpReq := &stock.GetLimitUpPoolRequest{
		Date: req.Date,
	}
	limitUpResp, err := rpc.StockClient.GetLimitUpPool(ctx, limitUpReq)
	if err != nil {
		log.Printf("获取涨停池失败: %v", err)
		return nil, err
	}

	// 3. 获取龙虎榜数据
	dtReq := &stock.GetDragonTigerListRequest{
		Date: req.Date,
	}
	dtResp, err := rpc.StockClient.GetDragonTigerList(ctx, dtReq)
	if err != nil {
		log.Printf("获取龙虎榜失败: %v", err)
		// 不要让整个请求失败，只需记录日志并传递 nil/空值
		dtResp = &stock.GetDragonTigerListResponse{}
	}

	// 4. 调用 LLM 提供商
	review, err := s.analyst.ReviewMarket(ctx, sectorResp.Sectors, limitUpResp.Stocks, dtResp.Items, req.Date)
	if err != nil {
		log.Printf("生成市场复盘失败: %v", err)
		return nil, err
	}

	return review, nil
}

// AnalyzeMarket implements the AIServiceImpl interface.
func (s *AIServiceImpl) AnalyzeMarket(ctx context.Context, req *ai.MarketAnalysisRequest) (resp *ai.MarketAnalysisResponse, err error) {
	log.Printf("收到市场分析请求: 日期=%s", req.Date)

	// 1. 获取板块数据
	sectorReq := &stock.GetMarketSectorsRequest{
		Type:  "concept",
		Limit: 20,
	}
	sectorResp, err := rpc.StockClient.GetMarketSectors(ctx, sectorReq)
	if err != nil {
		log.Printf("获取市场板块失败: %v", err)
		return nil, err
	}

	// 2. 获取涨停数据
	limitUpReq := &stock.GetLimitUpPoolRequest{
		Date: req.Date,
	}
	limitUpResp, err := rpc.StockClient.GetLimitUpPool(ctx, limitUpReq)
	if err != nil {
		log.Printf("获取涨停池失败: %v", err)
		return nil, err
	}

	// 3. 获取龙虎榜数据
	dtReq := &stock.GetDragonTigerListRequest{
		Date: req.Date,
	}
	dtResp, err := rpc.StockClient.GetDragonTigerList(ctx, dtReq)
	if err != nil {
		log.Printf("获取龙虎榜失败: %v", err)
		// 不要让整个请求失败，只需记录日志并传递 nil/空值
		dtResp = &stock.GetDragonTigerListResponse{}
	}

	// 4. 调用 LLM 提供商 (Analyst Agent - 初筛)
	// 4. 调用 LLM 提供商
	analysis, recStocks, err := s.analyst.AnalyzeMarket(ctx, sectorResp.Sectors, limitUpResp.Stocks, dtResp.Items, req.Date)
	if err != nil {
		log.Printf("生成市场分析失败: %v", err)
		return nil, err
	}

	// 5. 验证并修正推荐股票名称 (防止 AI 幻觉)
	for i, rec := range analysis.RecommendedStocks {
		// 预期格式: "股票名称 (代码): 理由"
		rec = strings.TrimSpace(rec)
		colon := strings.IndexAny(rec, ":：")
		if colon == -1 {
			continue
		}

		pre := strings.TrimSpace(rec[:colon])
		reason := strings.TrimSpace(rec[colon+1:])

		pre = strings.ReplaceAll(pre, "（", "(")
		pre = strings.ReplaceAll(pre, "）", ")")

		start := strings.LastIndex(pre, "(")
		end := strings.LastIndex(pre, ")")
		if start == -1 || end == -1 || end <= start {
			continue
		}

		currentName := strings.TrimSpace(pre[:start])
		code := strings.TrimSpace(pre[start+1 : end])
		code = strings.Trim(code, "()")

		realtimeResp, err := rpc.StockClient.GetRealtime(ctx, &stock.GetRealtimeRequest{Code: code})
		if err != nil || realtimeResp == nil || realtimeResp.Stock == nil || realtimeResp.Stock.Name == "" {
			log.Printf("无法验证股票代码: %s, err: %v", code, err)
			continue
		}

		realName := realtimeResp.Stock.Name
		if currentName != "" && currentName != realName {
			log.Printf("修正股票名称幻觉: 代码=%s, AI称=%s, 实际=%s", code, currentName, realName)
			analysis.RecommendedStocks[i] = fmt.Sprintf("%s (%s): %s", realName, code, reason)
		}
	}
	// 自动追踪推荐股票
	if len(recStocks) > 0 {
		go func(codes []string) {
			for _, code := range codes {
				// 调用 GetPrediction 进行详细分析和追踪
				log.Printf("对推荐股票 %s 进行详细预测...", code)
				_, err := s.GetPrediction(context.Background(), &ai.GetPredictionRequest{
					Code:  code,
					Days:  3,                // 预测未来3天
					Model: "glm-4.6v-flash", // 使用默认模型
				})
				if err != nil {
					log.Printf("自动预测失败: %s, %v", code, err)
				}
			}
		}(recStocks)
	}

	return analysis, nil
}

// ProcessMarketTrends implements the AIServiceImpl interface.
func (s *AIServiceImpl) ProcessMarketTrends(ctx context.Context, req *ai.ProcessMarketTrendsRequest) (resp *ai.ProcessMarketTrendsResponse, err error) {
	log.Printf("收到市场趋势处理请求: 条目数=%d", len(req.Items))

	items, err := s.analyst.ProcessMarketTrends(ctx, req.Items)
	if err != nil {
		log.Printf("处理市场趋势失败: %v", err)
		return nil, err
	}

	return &ai.ProcessMarketTrendsResponse{
		Items: items,
	}, nil
}
