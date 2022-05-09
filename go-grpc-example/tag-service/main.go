package main

import (
	"flag"
	"log"
	"net"

	pb "github.com/tag-service/proto"
	"github.com/tag-service/server"
	"google.golang.org/grpc"
)

var port string

func init() {
	flag.StringVar(&port, "p", "8000", "启动端口号")
	flag.Parse()
}

func main() {
	s := grpc.NewServer()
	pb.RegisterTagServiceServer(s, server.NewTagServer())

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("net.Listen err: %v", err)
	}

	err = s.Serve(lis)
	if err != nil {
		log.Fatal("server.Serve err: %v", err)
	}
}
