package main

import "go-product/RabbitMQ"

func main() {
	peileiTwo := RabbitMQ.NewRabbitMQRouting("exPeilei", "peilei_two")
	peileiTwo.Consume()
}
