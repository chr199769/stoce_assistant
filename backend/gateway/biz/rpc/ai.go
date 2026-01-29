package rpc

import (
	"sync"

	"stock_assistant/backend/gateway/config"
	"stock_assistant/backend/gateway/kitex_gen/ai/aiservice"
)

var (
	AIClient aiservice.Client
	aiOnce   sync.Once
)

func InitAI() {
	aiOnce.Do(func() {
		initAIClient()
	})
}

func initAIClient() {
	var err error
	addr := "127.0.0.1:8889"
	cfg := config.Get()
	if cfg != nil && cfg.RPC != nil && cfg.RPC.AIServiceAddr != "" {
		addr = cfg.RPC.AIServiceAddr
	}
	opts := getClientOptions(addr)
	AIClient, err = aiservice.NewClient("ai_service", opts...)
	if err != nil {
		panic(err)
	}
}
