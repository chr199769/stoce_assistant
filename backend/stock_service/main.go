package main

import (
	"log"
	"net"
	"github.com/cloudwego/kitex/server"
	stock "stock_assistant/backend/stock_service/kitex_gen/stock/stockservice"
	"stock_assistant/backend/stock_service/dal/mysql"
	"stock_assistant/backend/stock_service/dal/redis"
)

func main() {
	// Init Data Access Layer
	mysql.Init()
	redis.Init()

	addr, _ := net.ResolveTCPAddr("tcp", ":8888")
	svr := stock.NewServer(NewStockServiceImpl(), server.WithServiceAddr(addr))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
