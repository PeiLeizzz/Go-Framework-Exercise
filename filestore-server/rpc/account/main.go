package main

import (
	"filestore-server/rpc/account/handler"
	"filestore-server/rpc/account/proto"
	"github.com/go-micro/plugins/v4/registry/consul"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"log"
	"time"
)

func main() {
	reg := consul.NewRegistry(func(op *registry.Options) {
		op.Addrs = []string{
			"127.0.0.1:8500",
		}
	})
	service := micro.NewService(
		micro.Name("go.micro.service.user"),
		micro.RegisterTTL(10*time.Second),
		micro.RegisterInterval(5*time.Second),
		micro.Registry(reg),
	)
	// 解析命令行参数
	service.Init()

	proto.RegisterUserServiceHandler(service.Server(), new(handler.User))
	err := service.Run()
	if err != nil {
		log.Fatal(err.Error())
	}
}
