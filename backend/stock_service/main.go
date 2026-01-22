package main

import (
	"context"
	"log"
	"net"
	"stock_assistant/backend/stock_service/biz/provider/langfuse"
	"stock_assistant/backend/stock_service/biz/provider/sina"
	"stock_assistant/backend/stock_service/biz/worker"
	"stock_assistant/backend/stock_service/config"
	"stock_assistant/backend/stock_service/dal/model"
	"stock_assistant/backend/stock_service/dal/mysql"
	"stock_assistant/backend/stock_service/dal/redis"
	stock "stock_assistant/backend/stock_service/kitex_gen/stock/stockservice"

	"github.com/cloudwego/kitex/server"
	"github.com/joho/godotenv"
)

func backfillStockNames() {
	if mysql.DB == nil {
		return
	}
	var evals []model.EvaluationRecord
	// 查找股票名称为空的记录
	if err := mysql.DB.Where("stock_name = '' OR stock_name IS NULL").Find(&evals).Error; err != nil {
		log.Printf("查找需要回填的记录失败: %v", err)
		return
	}

	if len(evals) == 0 {
		return
	}

	log.Printf("正在回填 %d 条记录...", len(evals))
	client := sina.NewClient()
	ctx := context.Background()

	for _, e := range evals {
		info, err := client.GetStockInfo(ctx, e.StockCode)
		if err == nil {
			e.StockName = info.Name
			mysql.DB.Save(&e)
		} else {
			log.Printf("获取 %s 信息失败: %v", e.StockCode, err)
		}
	}
	log.Println("回填完成")
}

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

	// 回填股票名称
	backfillStockNames()
	// 初始化 Langfuse
	if err := langfuse.InitLangfuse(config.Get().Langfuse); err != nil {
		log.Printf("初始化 Langfuse 失败: %v", err)
	} else {
		log.Println("Langfuse 已初始化")
	}

	// 启动评估 Worker
	evalWorker := worker.NewEvalWorker()
	evalWorker.Start()

	addr, _ := net.ResolveTCPAddr("tcp", ":8888")
	svr := stock.NewServer(NewStockServiceImpl(), server.WithServiceAddr(addr))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
