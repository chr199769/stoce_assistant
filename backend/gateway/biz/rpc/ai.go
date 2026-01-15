package rpc

import (
	"sync"
	"time"

	"stock_assistant/backend/gateway/kitex_gen/ai/aiservice"

	"github.com/cloudwego/kitex/client"
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
	AIClient, err = aiservice.NewClient("ai_service",
		client.WithHostPorts("127.0.0.1:8889"),
		client.WithRPCTimeout(60*time.Second),
		client.WithConnectTimeout(3*time.Second),
	)
	if err != nil {
		panic(err)
	}
}
