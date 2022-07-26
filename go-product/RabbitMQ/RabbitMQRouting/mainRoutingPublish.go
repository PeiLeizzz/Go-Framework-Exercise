package main

import (
	"fmt"
	"go-product/RabbitMQ"
	"strconv"
	"time"
)

func main() {
	peileiOne := RabbitMQ.NewRabbitMQRouting("exPeilei", "peilei_one")
	peileiTwo := RabbitMQ.NewRabbitMQRouting("exPeilei", "peilei_two")
	for i := 0; i <= 10; i++ {
		peileiOne.Publish("Hello peilei one!" + strconv.Itoa(i))
		peileiTwo.Publish("Hello peilei two!" + strconv.Itoa(i))
		time.Sleep(1 * time.Second)
		fmt.Println(i)
	}
}
