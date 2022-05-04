package main

import (
	"encoding/json"
	"fmt"
	"geerpc"
	"geerpc/codec"
	"log"
	"net"
	"time"
)

func startServer1(addr chan string) {
	// 使用 :0 自动选取一个空闲端口
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("netword error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	geerpc.Accept(l)
}

func main1() {
	addr := make(chan string)
	// 启动服务端
	go startServer1(addr)

	// 客户端部分
	// 确保服务端端口监听成功后客户端再发起请求
	conn, _ := net.Dial("tcp", <-addr)
	defer func() {
		_ = conn.Close()
	}()

	// send options
	_ = json.NewEncoder(conn).Encode(geerpc.DefaultOption)
	// option 是直接写入 conn 的，而 header｜body 是写入 buf 的
	// 如果下面写得太快，会不会导致 header｜body 被 json Decode 读取掉了？
	// (json Decode 会把 conn 的所有内容取出放入它的缓冲区进行读取)
	// 实验证明：会，如果读取太慢，写入太快，在 header｜body 写完后
	// 才开始读取 option，此时 json 会把 option｜header｜body 全部读走
	// 后续 readRequest 会被阻塞
	// 最好这里加一个 sleep，保证 option 先被处理掉
	time.Sleep(time.Second)

	cc := codec.NewGobCodec(conn)
	for i := 0; i < 5; i++ {
		h := &codec.Header{
			ServiceMethod: "Foo.Sum",
			Seq:           uint64(i),
		}
		_ = cc.Write(h, fmt.Sprintf("geerpc req %d", h.Seq)) // request 写 conn
		// reponse 读 conn ...
		// reponse 写 conn ...
		_ = cc.ReadHeader(h) // request 读 conn
		var reply string
		_ = cc.ReadBody(&reply)
		log.Println("reply", reply)
	}

	// 总体流程可以总结为：
	// w -> gob -> bufio -> conn -> conn -> gob -> r
}
