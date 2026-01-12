package main

import (
	"log"
	"net"
	"github.com/cloudwego/kitex/server"
	stock "stock_assistant/backend/stock_service/kitex_gen/stock/stockservice"
)

func main() {
	addr, _ := net.ResolveTCPAddr("tcp", ":8888")
	svr := stock.NewServer(NewStockServiceImpl(), server.WithServiceAddr(addr))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
