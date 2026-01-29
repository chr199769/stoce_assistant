package rpc

import (
	"sync"

	"stock_assistant/backend/gateway/config"
	"stock_assistant/backend/knowledge_graph/kitex_gen/knowledge_graph/knowledgegraphservice"
)

var (
	KnowledgeGraphClient knowledgegraphservice.Client
	knowledgeGraphOnce   sync.Once
)

func InitKnowledgeGraph() {
	knowledgeGraphOnce.Do(func() {
		initKnowledgeGraphClient()
	})
}

func initKnowledgeGraphClient() {
	var err error
	addr := "127.0.0.1:8890"
	cfg := config.Get()
	if cfg != nil && cfg.RPC != nil && cfg.RPC.KnowledgeGraphAddr != "" {
		addr = cfg.RPC.KnowledgeGraphAddr
	}
	opts := getClientOptions(addr)
	KnowledgeGraphClient, err = knowledgegraphservice.NewClient("knowledge_graph", opts...)
	if err != nil {
		panic(err)
	}
}
