package main

import (
	"log"
	"net"
	"stock_assistant/backend/ai_service/biz/provider/langfuse"
	"stock_assistant/backend/ai_service/biz/provider/prompt"
	"stock_assistant/backend/ai_service/config"
	ai "stock_assistant/backend/ai_service/kitex_gen/ai/aiservice"

	"github.com/cloudwego/kitex/server"
)

func main() {
	// 初始化配置
	if err := config.Init(); err != nil {
		log.Printf("初始化配置失败: %v", err)
	}

	// 启动盘中监控
	// sentinel := worker.NewIntradaySentinel()
	// sentinel.Start()
	// defer sentinel.Stop()

	// 初始化 Langfuse
	if err := langfuse.InitLangfuse(config.Get().Langfuse); err != nil {
		log.Printf("初始化 Langfuse 失败: %v", err)
	} else {
		log.Println("Langfuse 已初始化")
	}

	// 初始化提示词管理器
	prompt.Init(langfuse.GetLangfuse())

	addr, _ := net.ResolveTCPAddr("tcp", ":8889")
	svr := ai.NewServer(NewAIServiceImpl(config.Get().LLMConfig), server.WithServiceAddr(addr))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
