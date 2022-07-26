package main

import (
	"log"
	"net/http"
	"sync"
)

// 目前已抢购的数量
var sum int64 = 0

// 预存商品数量
var productNum int64 = 10000

var mu sync.Mutex

// 获取秒杀商品
func GetOneProduct() bool {
	mu.Lock()
	defer mu.Unlock()
	if sum < productNum {
		sum += 1
		return true
	}
	return false
}

func GetProduct(w http.ResponseWriter, req *http.Request) {
	if GetOneProduct() {
		w.Write([]byte("true"))
	}
	w.Write([]byte("false"))
}

func main() {
	http.HandleFunc("/getOne", GetProduct)
	err := http.ListenAndServe(":8084", nil)
	if err != nil {
		log.Fatal(err)
	}
}
