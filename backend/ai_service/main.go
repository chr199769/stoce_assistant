package main

import (
	"log"
	"net"
	"github.com/cloudwego/kitex/server"
	ai "stock_assistant/backend/ai_service/kitex_gen/ai/aiservice"
)

func main() {
	addr, _ := net.ResolveTCPAddr("tcp", ":8889")
	svr := ai.NewServer(NewAIServiceImpl(), server.WithServiceAddr(addr))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
