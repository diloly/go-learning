package main

import (
	"log"
	"net"

	"github.com/cloudwego/kitex/server"
	"user/kitex_gen/user/userservice"
)

func main() {
	addr, _ := net.ResolveTCPAddr("tcp", ":9001")
	svr := userservice.NewServer(new(UserServiceImpl), server.WithServiceAddr(addr))
	log.Printf("🚀 用户服务(Kitex)启动: :%d\n", 9001)
	log.Fatal(svr.Run())
}
