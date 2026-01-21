package main

import (
	"log"
	"net"
	"github.com/cloudwego/kitex/server"
	stock "stock_assistant/backend/stock_service/kitex_gen/stock/stockservice"
	"stock_assistant/backend/stock_service/config"
	"stock_assistant/backend/stock_service/dal/mysql"
	"stock_assistant/backend/stock_service/dal/redis"
	"stock_assistant/backend/stock_service/biz/worker"
	"stock_assistant/backend/stock_service/biz/provider/langfuse"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	// Init Config
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to init config: %v", err)
	}

	// Init Data Access Layer
	mysql.Init()
	redis.Init()
	// Init Langfuse
	if err := langfuse.InitLangfuse(config.Get().Langfuse); err != nil {
		log.Printf("Failed to init Langfuse: %v", err)
	} else {
		log.Println("Langfuse initialized")
	}
	
	// Start EvalWorker
	evalWorker := worker.NewEvalWorker()
	evalWorker.Start()

	addr, _ := net.ResolveTCPAddr("tcp", ":8888")
	svr := stock.NewServer(NewStockServiceImpl(), server.WithServiceAddr(addr))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
