package main

import (
	"context"
	"fmt"
	"geerpc"
	"log"
	"sync"
	"time"
)

func main2() {
	log.SetFlags(0)
	addr := make(chan string)
	// 服务器可以并发接收多个连接
	// 对于每个连接，可以并发接收多个请求、处理并发送多个响应（加锁）
	go startServer1(addr)
	client, _ := geerpc.Dial("tcp", <-addr)
	defer func() {
		_ = client.Close()
	}()

	// 如 Day1 启动所述，写得太快会使解析 option 时把 header｜body 的内容读取了
	// 这边增加一个延时，减少这种情况的发生
	time.Sleep(time.Second)
	// send request & receive response
	var wg sync.WaitGroup
	// 客户端可以并发发送多个请求（加锁）
	// 但客户端接收消息是单 goroutine 的（receive()）
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := fmt.Sprintf("geerpc req %d", i)
			var reply string
			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Println("reply:", reply)
		}(i)
	}
	wg.Wait()
}
