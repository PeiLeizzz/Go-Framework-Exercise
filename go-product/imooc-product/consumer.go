package main

import (
	"fmt"
	"go-product/RabbitMQ"
	"go-product/imooc-product/common"
	"go-product/imooc-product/repositories"
	"go-product/imooc-product/services"
)

func main() {
	db, err := common.NewMysqlConn()
	if err != nil {
		fmt.Println(err)
	}

	product := repositories.NewProductManager("product", db)
	productService := services.NewProductService(product)
	order := repositories.NewOrderManager("order", db)
	orderService := services.NewOrderService(order)

	rabbitmqConsumerSimple := RabbitMQ.NewRabbitMQSimple("imoocProduct")
	rabbitmqConsumerSimple.ConsumeOrder(orderService, productService)
}
