package main

import (
	"log"
	"net"
	"os"
	stock_rpc "stock_assistant/backend/stock_service/biz/rpc"
	"stock_assistant/backend/stock_service/biz/worker"
	"stock_assistant/backend/stock_service/config"
	"stock_assistant/backend/stock_service/dal/mysql"
	"stock_assistant/backend/stock_service/dal/redis"
	stock "stock_assistant/backend/stock_service/kitex_gen/stock/stockservice"
	"stock_assistant/backend/common/langfuse"

	"github.com/cloudwego/kitex/server"
)

func main() {
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

	newsKGWorker := worker.NewNewsKGWorker()
	if newsKGWorker != nil {
		newsKGWorker.Start()
		defer newsKGWorker.Stop()
	}
	snapshotKGWorker := worker.NewSnapshotKGWorker()
	if snapshotKGWorker != nil {
		snapshotKGWorker.Start()
		defer snapshotKGWorker.Stop()
	}

	addrValue := ":8888"
	if v := os.Getenv("STOCK_SERVICE_ADDR"); v != "" {
		addrValue = v
	}
	addr, _ := net.ResolveTCPAddr("tcp", addrValue)
	svr := stock.NewServer(NewStockServiceImpl(), server.WithServiceAddr(addr))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
