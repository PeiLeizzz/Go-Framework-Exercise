package main

import (
	"fmt"
	"go-product/RabbitMQ"
	"strconv"
	"time"
)

func main() {
	rabbitmq := RabbitMQ.NewRabbitMQSimple("peileiWork")

	for i := 0; i <= 100; i++ {
		rabbitmq.Publish("hello peilei work!" + strconv.Itoa(i))
		time.Sleep(1 * time.Second)
		fmt.Println(i)
	}
}
