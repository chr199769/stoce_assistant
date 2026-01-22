package main

import (
	"context"
	"log"
	"os"
	"stock_assistant/backend/ai_service/biz/provider/llm"
	ai "stock_assistant/backend/ai_service/kitex_gen/ai"
	"stock_assistant/backend/ai_service/kitex_gen/stock"
	"stock_assistant/backend/ai_service/kitex_gen/stock/stockservice"
	"strings"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/google/uuid"
)

// AIServiceImpl implements the last service interface defined in the IDL.
type AIServiceImpl struct {
	llmProvider llm.Provider
	stockClient stockservice.Client
}

func NewAIServiceImpl(llmConfig *llm.FileConfig) *AIServiceImpl {
	stockAddr := os.Getenv("STOCK_SERVICE_ADDR")
	if stockAddr == "" {
		stockAddr = "localhost:8888"
	}
	c, err := stockservice.NewClient("stock_service", client.WithHostPorts(stockAddr))
	if err != nil {
		log.Printf("初始化 stock 客户端失败: %v", err)
	}

	p, err := llm.NewLangChainProvider(context.Background(), c, llmConfig)
	if err != nil {
		log.Printf("初始化 langchain provider 失败: %v", err)
		// 如果我们严格不希望使用 mock，这里应该 panic。
		// 所以让我们直接 log fatal。
		log.Fatalf("严重错误: 初始化 LLM provider 失败且 mock 未启用: %v", err)
	}

	return &AIServiceImpl{
		llmProvider: p,
		stockClient: c,
	}
}

// GetPrediction implements the AIServiceImpl interface.
func (s *AIServiceImpl) GetPrediction(ctx context.Context, req *ai.GetPredictionRequest) (resp *ai.GetPredictionResponse, err error) {
	log.Printf("收到预测请求: 代码=%s, 模型=%s", req.Code, req.Model)
	analysis, confidence, newsSummary, traceID, err := s.llmProvider.Predict(ctx, req.Code, req.Days, req.Model)
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
			Id:             uuid.New().String(),
			StockCode:      req.Code,
			PredictionDate: time.Now().Format("2006-01-02 15:04:05"),
			Content:        analysis,
			Confidence:     confidence,
			Trend:          trend,
			TraceId:        traceID,
		},
	}

	_, saveErr := s.stockClient.SavePrediction(ctx, saveReq)
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

	stocks, err := s.llmProvider.RecognizeImage(ctx, req.ImageData, req.Model)
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
	sectorResp, err := s.stockClient.GetMarketSectors(ctx, sectorReq)
	if err != nil {
		log.Printf("获取市场板块失败: %v", err)
		return nil, err
	}

	// 2. 获取涨停数据
	limitUpReq := &stock.GetLimitUpPoolRequest{
		Date: req.Date,
	}
	limitUpResp, err := s.stockClient.GetLimitUpPool(ctx, limitUpReq)
	if err != nil {
		log.Printf("获取涨停池失败: %v", err)
		return nil, err
	}

	// 3. 获取龙虎榜数据
	dtReq := &stock.GetDragonTigerListRequest{
		Date: req.Date,
	}
	dtResp, err := s.stockClient.GetDragonTigerList(ctx, dtReq)
	if err != nil {
		log.Printf("获取龙虎榜失败: %v", err)
		// 不要让整个请求失败，只需记录日志并传递 nil/空值
		dtResp = &stock.GetDragonTigerListResponse{}
	}

	// 4. 调用 LLM 提供商
	review, err := s.llmProvider.ReviewMarket(ctx, sectorResp.Sectors, limitUpResp.Stocks, dtResp.Items, req.Date)
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
	sectorResp, err := s.stockClient.GetMarketSectors(ctx, sectorReq)
	if err != nil {
		log.Printf("获取市场板块失败: %v", err)
		return nil, err
	}

	// 2. 获取涨停数据
	limitUpReq := &stock.GetLimitUpPoolRequest{
		Date: req.Date,
	}
	limitUpResp, err := s.stockClient.GetLimitUpPool(ctx, limitUpReq)
	if err != nil {
		log.Printf("获取涨停池失败: %v", err)
		return nil, err
	}

	// 3. 获取龙虎榜数据
	dtReq := &stock.GetDragonTigerListRequest{
		Date: req.Date,
	}
	dtResp, err := s.stockClient.GetDragonTigerList(ctx, dtReq)
	if err != nil {
		log.Printf("获取龙虎榜失败: %v", err)
		// 不要让整个请求失败，只需记录日志并传递 nil/空值
		dtResp = &stock.GetDragonTigerListResponse{}
	}

	// 4. 调用 LLM 提供商
	log.Printf("调用 AnalyzeMarket 参数: 板块数=%d, 涨停数=%d, 龙虎榜数=%d", len(sectorResp.Sectors), len(limitUpResp.Stocks), len(dtResp.Items))
	analysis, err := s.llmProvider.AnalyzeMarket(ctx, sectorResp.Sectors, limitUpResp.Stocks, dtResp.Items, req.Date)
	if err != nil {
		log.Printf("生成市场分析失败: %v", err)
		return nil, err
	}

	return analysis, nil
}
