package main

import (
	"context"
	"geerpc"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

func startServer5(addr chan string) {
	var foo Foo
	l, _ := net.Listen("tcp", ":9999")
	_ = geerpc.Register(&foo)
	geerpc.HandleHTTP()
	addr <- l.Addr().String()
	_ = http.Serve(l, nil)
}

func call(addrCh chan string) {
	client, _ := geerpc.DialHTTP("tcp", <-addrCh)
	defer func() {
		_ = client.Close()
	}()

	time.Sleep(time.Second)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{Num1: i, Num2: i * i}
			var reply int
			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)
	}
	wg.Wait()
}

func call2(addrCh chan string) {
	// 这里虽然不会出错，可以正常将 options 发给 server，
	// 但其实 server 那边也无法处理 option，因为处理 option 的逻辑在 ServeConn 中
	// 而 ServeConn 是在建立完 HTTP 连接后才调用的

	// 同样下面的 Call，server 无法正常处理（客户端接收响应时会收到无法解析的响应）
	// 因为 server 是用 HandleHTTP() 开启的
	// 需要通过 http 请求，才能被其 ServeHTTP 函数处理
	// 而 RPC Conn 是在 ServeHTTP 函数内部建立的
	// 如果不通过 HTTP 请求，就无法建立 RPC 连接
	client, _ := geerpc.Dial("tcp", <-addrCh)
	defer func() {
		_ = client.Close()
	}()

	time.Sleep(time.Second)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{Num1: i, Num2: i * i}
			var reply int
			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)
	}
	wg.Wait()
}

func main() {
	log.SetFlags(0)
	ch := make(chan string)
	go call(ch)
	// go call2(ch)
	startServer5(ch)
}
