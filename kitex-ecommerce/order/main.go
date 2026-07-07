package main

import (
	"log"
	"net"

	"github.com/cloudwego/kitex/server"
	"order/kitex_gen/order/orderservice"
)

func main() {
	startOrderConsumer()

	addr, _ := net.ResolveTCPAddr("tcp", ":9003")
	svr := orderservice.NewServer(new(OrderServiceImpl), server.WithServiceAddr(addr))
	log.Printf("🚀 订单服务(Kitex+RocketMQ)启动: :%d\n", 9003)
	log.Fatal(svr.Run())
}
