package main

import (
	"context"
	"log"
	"net"
	"stock_assistant/backend/knowledge_graph/config"
	"stock_assistant/backend/knowledge_graph/dal/graphdb"
	knowledge_graph "stock_assistant/backend/knowledge_graph/kitex_gen/knowledge_graph/knowledgegraphservice"
	"time"

	"github.com/cloudwego/kitex/server"
)

func main() {
	if err := config.Init(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	if err := graphdb.Init(); err != nil {
		log.Fatalf("Neo4j 初始化失败: %v", err)
	}

	startCleanup()

	addrValue := ":8890"
	cfg := config.Get()
	if cfg != nil && cfg.Server != nil && cfg.Server.Addr != "" {
		addrValue = cfg.Server.Addr
	}
	addr, _ := net.ResolveTCPAddr("tcp", addrValue)
	svr := knowledge_graph.NewServer(new(KnowledgeGraphServiceImpl), server.WithServiceAddr(addr))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}

func startCleanup() {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			cleanupIntradaySummaries()
			<-ticker.C
		}
	}()
}

func cleanupIntradaySummaries() {
	before := time.Now().AddDate(0, 0, -30).Unix()
	_, _ = graphdb.DeleteEventsBySourceBefore(contextBackground(), "intraday_summary", before)
}

func contextBackground() context.Context { return context.Background() }
