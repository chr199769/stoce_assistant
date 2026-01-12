package llm

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type Provider interface {
	Predict(ctx context.Context, stockCode string, days int32, modelName string) (string, float64, string, error)
}

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (p *MockProvider) Predict(ctx context.Context, stockCode string, days int32, modelName string) (string, float64, string, error) {
	// Simulate LLM delay
	time.Sleep(1 * time.Second)

	// Mock prediction logic
	rand.Seed(time.Now().UnixNano())
	confidence := 0.7 + rand.Float64()*0.2
	
	trend := "上涨"
	if rand.Float64() < 0.5 {
		trend = "下跌"
	}

	analysis := fmt.Sprintf("基于%s过去30天的技术指标分析，MACD出现金叉，KDJ指标显示超卖。结合近期行业利好政策，预计未来%d天股价将呈现%s趋势。", stockCode, days, trend)
	newsSummary := "1. 行业龙头发布利好财报\n2. 国家出台相关扶持政策\n3. 市场资金持续净流入"

	return analysis, confidence, newsSummary, nil
}
