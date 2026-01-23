package rpc

import (
	"log"
	"sync"

	"stock_assistant/backend/ai_service/kitex_gen/stock/stockservice"

	"github.com/cloudwego/kitex/client"
)

var (
	StockClient stockservice.Client
	once        sync.Once
)

func Init() {
	once.Do(func() {
		initStockClient()
	})
}

func initStockClient() {
	// TODO: 从配置读取 Stock Service 地址
	c, err := stockservice.NewClient("stock_service", client.WithHostPorts("localhost:8888"))
	if err != nil {
		log.Fatalf("[RPC] 初始化 Stock Client 失败: %v", err)
	}
	StockClient = c
}
