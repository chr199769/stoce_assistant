package main

import (
	"context"
	"fmt"

	"stock_assistant/backend/ai_service/biz/provider/llm/predictor"
	ai_rpc "stock_assistant/backend/ai_service/biz/rpc"
	"stock_assistant/backend/ai_service/config"
)

func main() {
	config.Init()
	ai_rpc.Init()
	p := predictor.NewProvider(config.Get().LLMConfig)
	ctx := context.Background()
	res, conf, summary, _, _, _, err := p.Predict(ctx, "sh600519", 5, "fractal")
	if err != nil {
		fmt.Printf("分形预测错误: %v\n", err)
		return
	}
	fmt.Printf("分形预测相似度: %.2f\n", conf)
	fmt.Println("分形预测摘要:")
	fmt.Println(summary)
	fmt.Println("分形预测分析:")
	fmt.Println(res)
}
