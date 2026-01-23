package rpc

import (
	"log"
	"sync"

	"stock_assistant/backend/stock_service/kitex_gen/ai/aiservice"

	"github.com/cloudwego/kitex/client"
)

var (
	AIClient aiservice.Client
	once     sync.Once
)

func Init() {
	once.Do(func() {
		initAIClient()
	})
}

func initAIClient() {
	// TODO: 从配置读取 AI Service 地址
	c, err := aiservice.NewClient("ai_service", client.WithHostPorts("localhost:8889"))
	if err != nil {
		log.Fatalf("[RPC] 初始化 AI Client 失败: %v", err)
	}
	AIClient = c
}
