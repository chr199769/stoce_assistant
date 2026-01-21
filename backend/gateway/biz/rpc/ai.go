package rpc

import (
	"sync"

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
	// In a real environment, use service discovery
	opts := getClientOptions("AI_SERVICE_ADDR", "127.0.0.1:8889")
	AIClient, err = aiservice.NewClient("ai_service", opts...)
	if err != nil {
		panic(err)
	}
}
