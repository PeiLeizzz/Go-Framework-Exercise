package main

import (
	"context"
	"flag"
	"io"
	"log"

	pb "github.com/go-grpc-example/proto"
	"google.golang.org/grpc"
)

var port string

func init() {
	flag.StringVar(&port, "p", "8000", "启动端口号")
	flag.Parse()
}

// 一次发送，一次接收
func SayHello(client pb.GreeterClient, r *pb.HelloRequest) error {
	// 调用远程方法
	resp, _ := client.SayHello(context.Background(), r)
	log.Printf("client.SayHello resp: %s", resp.Message)
	return nil
}

// 一次发送，多次接收
func SayList(client pb.GreeterClient, r *pb.HelloRequest) error {
	stream, _ := client.SayList(context.Background(), r)
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		log.Printf("resp: %v", resp)
	}
	return nil
}

// 多次发送，一次接收
func SayRecord(client pb.GreeterClient, r *pb.HelloRequest) error {
	stream, _ := client.SayRecord(context.Background())
	for n := 1; n <= 7; n++ {
		_ = stream.Send(r)
	}
	resp, _ := stream.CloseAndRecv() // 与 server 端的 stream.SendAndClose() 配套使用
	log.Printf("resp: %v", resp)
	return nil
}

func SayRoute(client pb.GreeterClient, r *pb.HelloRequest) error {
	stream, _ := client.SayRoute(context.Background())
	for n := 1; n <= 7; n++ {
		_ = stream.Send(r)
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		log.Printf("resp%d: %v", n, resp)
	}

	_ = stream.CloseSend()
	return nil
}

func main() {
	conn, _ := grpc.Dial(":"+port, grpc.WithInsecure())
	defer conn.Close()

	client := pb.NewGreeterClient(conn)
	r := &pb.HelloRequest{Name: "peilei"}
	_ = SayHello(client, r)
	_ = SayList(client, r)
	_ = SayRecord(client, r)
	_ = SayRoute(client, r)
}
