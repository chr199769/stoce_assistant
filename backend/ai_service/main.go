package main

import (
	"log"
	"net"
	"stock_assistant/backend/ai_service/biz/worker"
	"stock_assistant/backend/ai_service/biz/provider/llm"
	ai "stock_assistant/backend/ai_service/kitex_gen/ai/aiservice"

	"github.com/cloudwego/kitex/server"
)

func main() {
	// Start Intraday Sentinel
	sentinel := worker.NewIntradaySentinel()
	sentinel.Start()
	defer sentinel.Stop()

	// Init Langfuse
	if err := llm.InitLangfuse(); err != nil {
		log.Printf("Failed to init Langfuse: %v", err)
	} else {
		log.Println("Langfuse initialized")
	}

	addr, _ := net.ResolveTCPAddr("tcp", ":8889")
	svr := ai.NewServer(NewAIServiceImpl(), server.WithServiceAddr(addr))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
