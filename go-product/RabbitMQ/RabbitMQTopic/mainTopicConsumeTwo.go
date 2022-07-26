package main

import "go-product/RabbitMQ"

func main() {
	peileiTwo := RabbitMQ.NewRabbitMQTopic("exPeileiTopic", "peilei.*.two")
	peileiTwo.Consume()
}
