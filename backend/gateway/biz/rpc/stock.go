package rpc

import (
	"sync"

	"stock_assistant/backend/gateway/config"
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
	addr := "127.0.0.1:8888"
	cfg := config.Get()
	if cfg != nil && cfg.RPC != nil && cfg.RPC.StockServiceAddr != "" {
		addr = cfg.RPC.StockServiceAddr
	}
	opts := getClientOptions(addr)
	StockClient, err = stockservice.NewClient("stock_service", opts...)
	if err != nil {
		panic(err)
	}
}
