package main

import (
	"fmt"
	"go-product/RabbitMQ"
)

func main() {
	rabbitmq := RabbitMQ.NewRabbitMQSimple("peileiSimple")
	rabbitmq.Publish("hello peilei simple!")
	fmt.Println("发送成功！")
}
