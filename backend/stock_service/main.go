package main

import (
	"log"
	"net"
	"stock_assistant/backend/stock_service/biz/provider/langfuse"
	stock_rpc "stock_assistant/backend/stock_service/biz/rpc"
	"stock_assistant/backend/stock_service/biz/worker"
	"stock_assistant/backend/stock_service/config"
	"stock_assistant/backend/stock_service/dal/mysql"
	"stock_assistant/backend/stock_service/dal/redis"
	stock "stock_assistant/backend/stock_service/kitex_gen/stock/stockservice"

	"github.com/cloudwego/kitex/server"
	"github.com/joho/godotenv"
)

func main() {
	// 加载 .env 文件（如果存在）
	if err := godotenv.Load(); err != nil {
		log.Println("未找到 .env 文件，使用环境变量")
	}

	// 初始化配置
	if err := config.Init(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	// 初始化数据访问层
	mysql.Init()
	redis.Init()
	stock_rpc.Init()

	// 初始化 Langfuse
	if err := langfuse.InitLangfuse(config.Get().Langfuse); err != nil {
		log.Printf("初始化 Langfuse 失败: %v", err)
	} else {
		log.Println("Langfuse 已初始化")
	}

	// 启动评估 Worker
	evalWorker := worker.NewEvalWorker()
	evalWorker.Start()

	// 启动市场趋势分析 Worker
	trendWorker := worker.NewTrendWorker()
	if trendWorker != nil {
		trendWorker.Start()
		defer trendWorker.Stop()
	}

	addr, _ := net.ResolveTCPAddr("tcp", ":8888")
	svr := stock.NewServer(NewStockServiceImpl(), server.WithServiceAddr(addr))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
