package rpc

import (
	"log"
	"sync"

	"stock_assistant/backend/ai_service/config"
	"stock_assistant/backend/ai_service/kitex_gen/stock/stockservice"
	knowledge_graph "stock_assistant/backend/knowledge_graph/kitex_gen/knowledge_graph/knowledgegraphservice"

	"github.com/cloudwego/kitex/client"
)

var (
	StockClient stockservice.Client
	KnowledgeGraphClient knowledge_graph.Client
	once                   sync.Once
)

func Init() {
	once.Do(func() {
		initStockClient()
		initKnowledgeGraphClient()
	})
}

func initStockClient() {
	addr := "localhost:8888"
	cfg := config.Get()
	if cfg != nil && cfg.RPC != nil && cfg.RPC.StockServiceAddr != "" {
		addr = cfg.RPC.StockServiceAddr
	}
	c, err := stockservice.NewClient("stock_service", client.WithHostPorts(addr))
	if err != nil {
		log.Fatalf("[RPC] 初始化 Stock Client 失败: %v", err)
	}
	StockClient = c
}

func initKnowledgeGraphClient() {
	addr := "localhost:8890"
	cfg := config.Get()
	if cfg != nil && cfg.RPC != nil && cfg.RPC.KnowledgeGraphAddr != "" {
		addr = cfg.RPC.KnowledgeGraphAddr
	}
	c, err := knowledge_graph.NewClient("knowledge_graph", client.WithHostPorts(addr))
	if err != nil {
		log.Fatalf("[RPC] 初始化 Knowledge Graph Client 失败: %v", err)
	}
	KnowledgeGraphClient = c
}
