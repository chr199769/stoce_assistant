package rpc

import (
	"sync"

	"stock_assistant/backend/gateway/kitex_gen/stock/stockservice"
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
	opts := getClientOptions("STOCK_SERVICE_ADDR", "127.0.0.1:8888")
	StockClient, err = stockservice.NewClient("stock_service", opts...)
	if err != nil {
		panic(err)
	}
}
