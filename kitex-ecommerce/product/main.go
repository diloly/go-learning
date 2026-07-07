package main

import (
	"log"
	"net"

	"github.com/cloudwego/kitex/server"
	"product/kitex_gen/product/productservice"
)

func main() {
	addr, _ := net.ResolveTCPAddr("tcp", ":9002")
	svr := productservice.NewServer(new(ProductServiceImpl), server.WithServiceAddr(addr))
	log.Printf("🚀 商品服务(Kitex)启动: :%d\n", 9002)
	log.Fatal(svr.Run())
}
