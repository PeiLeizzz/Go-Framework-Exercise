package main

import (
	"io"
	"log"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", ":2020")
	if err != nil {
		panic(err)
	}

	done := make(chan struct{})
	// 负责将 conn 中的消息复制到 stdout 中的协程
	// 如果复制过程中出错（io.Copy 退出），则告知主协程
	// done 的作用是防止主协程在还未复制完发来的消息时退出
	go func() {
		io.Copy(os.Stdout, conn) // NOTE: ignore errors
		log.Println("done")
		done <- struct{}{}
	}()

	// 负责将 stdin 中的消息复制到 conn 中
	mustCopy(conn, os.Stdin)
	conn.Close()
	<-done
}

func mustCopy(dst io.Writer, src io.Reader) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Fatal(err)
	}
}
