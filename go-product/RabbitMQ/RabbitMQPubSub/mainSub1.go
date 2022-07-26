package main

import "go-product/RabbitMQ"

func main() {
	rabbitmq := RabbitMQ.NewRabbitMQPubSub("peileiPubSub")
	rabbitmq.Consume()
}
