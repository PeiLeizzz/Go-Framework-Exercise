package main

import "go-product/RabbitMQ"

func main() {
	rabbitmq := RabbitMQ.NewRabbitMQSimple("peileiWork")
	rabbitmq.Consume()
}
