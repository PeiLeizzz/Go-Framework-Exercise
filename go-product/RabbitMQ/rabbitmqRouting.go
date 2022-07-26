package RabbitMQ

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

type RabbitMQRouting struct {
	*rabbitMQ
}

var _ RabbitMQCommon = (*RabbitMQRouting)(nil)

// 路由模式创建 RabbitMQ 实例
func NewRabbitMQRouting(exchangeName, routingKey string) *RabbitMQRouting {
	return &RabbitMQRouting{
		newRabbitMQ("", exchangeName, routingKey),
	}
}

// 路由模式生产消息
func (r *RabbitMQRouting) Publish(message string) {
	// 1. 尝试创建交换机
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an exchange")

	// 2. 发送消息
	err = r.channel.Publish(
		r.Exchange,
		r.Key, // 必须要 key
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
}

// 路由模式消费消息
func (r *RabbitMQRouting) Consume() {
	// 1. 尝试创建交换机
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an exchange")

	// 2. 尝试创建队列（不需要写队列名）
	q, err := r.channel.QueueDeclare(
		"", // 随机队列名称
		false,
		false,
		true,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare a queue")

	// 3. 绑定队列到 exchange 中
	err = r.channel.QueueBind(
		q.Name,
		r.Key,
		r.Exchange,
		false,
		nil,
	)

	// 4. 消费消息
	messages, err := r.channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	forever := make(chan bool)

	go func() {
		for d := range messages {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	fmt.Println("[*] Waiting for messages, To exit press CTRL+C")
	<-forever
}
