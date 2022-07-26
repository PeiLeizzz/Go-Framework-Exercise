package main

import (
	"fmt"
	"go-product/RabbitMQ"
	"strconv"
	"time"
)

func main() {
	peileiOne := RabbitMQ.NewRabbitMQTopic("exPeileiTopic", "peilei.topic.one")
	peileiTwo := RabbitMQ.NewRabbitMQTopic("exPeileiTopic", "peilei.topic.two")
	for i := 0; i <= 10; i++ {
		peileiOne.Publish("Hello peilei topic one!" + strconv.Itoa(i))
		peileiTwo.Publish("Hello peilei topic two!" + strconv.Itoa(i))
		time.Sleep(1 * time.Second)
		fmt.Println(i)
	}
}
