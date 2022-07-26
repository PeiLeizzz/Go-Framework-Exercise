package main

import "go-product/RabbitMQ"

func main() {
	peileiAll := RabbitMQ.NewRabbitMQTopic("exPeileiTopic", "#")
	peileiAll.Consume()
}
