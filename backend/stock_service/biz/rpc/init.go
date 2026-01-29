package rpc

import (
	"log"
	"sync"

	knowledgegraphservice "stock_assistant/backend/knowledge_graph/kitex_gen/knowledge_graph/knowledgegraphservice"
	"stock_assistant/backend/stock_service/config"
	"stock_assistant/backend/stock_service/kitex_gen/ai/aiservice"

	"github.com/cloudwego/kitex/client"
)

var (
	AIClient aiservice.Client
	KnowledgeGraphClient knowledgegraphservice.Client
	once     sync.Once
)

func Init() {
	once.Do(func() {
		initAIClient()
		initKnowledgeGraphClient()
	})
}

func initAIClient() {
	addr := "localhost:8889"
	cfg := config.Get()
	if cfg != nil && cfg.RPC != nil && cfg.RPC.AIServiceAddr != "" {
		addr = cfg.RPC.AIServiceAddr
	}
	c, err := aiservice.NewClient("ai_service", client.WithHostPorts(addr))
	if err != nil {
		log.Fatalf("[RPC] 初始化 AI Client 失败: %v", err)
	}
	AIClient = c
}

func initKnowledgeGraphClient() {
	addr := "localhost:8890"
	cfg := config.Get()
	if cfg != nil && cfg.RPC != nil && cfg.RPC.KnowledgeGraphAddr != "" {
		addr = cfg.RPC.KnowledgeGraphAddr
	}
	c, err := knowledgegraphservice.NewClient("knowledge_graph", client.WithHostPorts(addr))
	if err != nil {
		log.Fatalf("[RPC] 初始化 KnowledgeGraph Client 失败: %v", err)
	}
	KnowledgeGraphClient = c
}
