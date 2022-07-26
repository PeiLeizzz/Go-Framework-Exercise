package main

import (
	"filestore-server/rpc/upload/handler"
	proto "filestore-server/rpc/upload/proto"
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
		micro.Name("go.micro.service.upload"),
		micro.RegisterTTL(10*time.Second),
		micro.RegisterInterval(5*time.Second),
		micro.Registry(reg),
	)
	// 解析命令行参数
	service.Init()

	proto.RegisterUploadServiceHandler(service.Server(), new(handler.Upload))
	err := service.Run()
	if err != nil {
		log.Fatal(err.Error())
	}
}
