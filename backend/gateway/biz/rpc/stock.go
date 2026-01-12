package rpc

import (
	"sync"

	"github.com/cloudwego/kitex/client"
	"stock_assistant/backend/stock_service/kitex_gen/stock/stockservice"
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
	var err error
	// In a real environment, use service discovery (e.g., etcd/consul)
	// For local development/demo, direct address is fine or simple resolver
	StockClient, err = stockservice.NewClient("stock_service", client.WithHostPorts("127.0.0.1:8888"))
	if err != nil {
		panic(err)
	}
}
