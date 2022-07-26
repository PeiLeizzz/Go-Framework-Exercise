package main

import (
	"go-cache/fast"
	"strconv"
	"time"
)

type Value struct {
	A string
	B int
	C time.Time
	D []byte
	E float32
	F *string
	T T
}

type T struct {
	H int
	I int
	K int
	L int
	M int
	N int
}

func main() {
	cache := fast.NewFastCache(0, 1024, nil)

	for i := 0; i < 10000000; i++ {
		cache.Set(strconv.Itoa(i), &Value{})
	}

	for i := 0; ; i++ {
		cache.Del(strconv.Itoa(i))
		cache.Set(strconv.Itoa(10000000+i), &Value{})
		time.Sleep(5 * time.Millisecond)
	}
}
