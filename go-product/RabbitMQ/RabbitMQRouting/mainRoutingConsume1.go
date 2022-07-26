package main

import "go-product/RabbitMQ"

func main() {
	peileiOne := RabbitMQ.NewRabbitMQRouting("exPeilei", "peilei_one")
	peileiOne.Consume()
}