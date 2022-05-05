package main

import (
	"context"
	"geerpc"
	"log"
	"sync"
	"time"
)

func (f Foo) Timeout(args Args, reply *int) error {
	time.Sleep(time.Second * 2)
	return nil
}

func main() {
	log.SetFlags(0)
	addr := make(chan string)
	go startServer(addr)

	client, _ := geerpc.Dial("tcp", <-addr)
	defer func() {
		_ = client.Close()
	}()

	time.Sleep(time.Second)
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			args := &Args{Num1: 1, Num2: 1} // 也可以不用 &，不影响正确性
			ctx, _ := context.WithTimeout(context.Background(), time.Second)
			var reply int
			if err := client.Call(ctx, "Foo.Timeout", args, &reply); err != nil {
				log.Println("call Foo.Timeout error:", err)
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}()
	}

	time.Sleep(time.Second * 2)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{Num1: i, Num2: i * i} // 也可以不用 &，不影响正确性
			var reply int
			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)
	}
	wg.Wait()
}
