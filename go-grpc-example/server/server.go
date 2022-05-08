package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net"

	pb "github.com/go-grpc-example/proto"
	"google.golang.org/grpc"
)

type GreeterServer struct{}

// 一次接收，一次发送
func (s *GreeterServer) SayHello(ctx context.Context, r *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "hello.world"}, nil
}

// 一次接收，多次发送
func (s *GreeterServer) SayList(r *pb.HelloRequest, stream pb.Greeter_SayListServer) error {
	for n := 0; n <= 6; n++ {
		_ = stream.Send(&pb.HelloReply{Message: "hello.list"})
	}
	return nil
}

// 多次接收，一次发送
func (s *GreeterServer) SayRecord(stream pb.Greeter_SayRecordServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			message := &pb.HelloReply{Message: "say.record"}
			return stream.SendAndClose(message) // 与 Client 端的 stream.CloseAndRecv() 配套使用
		}
		if err != nil {
			return err
		}

		log.Printf("req: %v", req)
	}
}

func (s *GreeterServer) SayRoute(stream pb.Greeter_SayRouteServer) error {
	n := 0
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		n++
		log.Printf("req%d: %v", n, req)

		_ = stream.Send(&pb.HelloReply{Message: "say.route"})
	}
}

var port string

func init() {
	flag.StringVar(&port, "p", "8000", "启动端口号")
	flag.Parse()
}

func main() {
	server := grpc.NewServer()
	// 把 GreeterServer 注册到 gRPC Server 内部的注册中心
	pb.RegisterGreeterServer(server, &GreeterServer{})
	lis, _ := net.Listen("tcp", ":"+port)
	server.Serve(lis)
}
