package RabbitMQ

import (
	"github.com/streadway/amqp"
	"log"
)

type RabbitMQPubSub struct {
	*rabbitMQ
}

var _ RabbitMQCommon = (*RabbitMQPubSub)(nil)

// 订阅模式下创建 RabbitMQ 实例
func NewRabbitMQPubSub(exchangeName string) *RabbitMQPubSub {
	return &RabbitMQPubSub{
		newRabbitMQ("", exchangeName, ""),
	}
}

// 订阅模式下生产消息
func (r *RabbitMQPubSub) Publish(message string) {
	// 1. 尝试创建交换机
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		"fanout", // 交换机类型
		true,     // 是否持久化
		false,    // 是否自动删除
		false,    // 设置为 true 代表这个 exchange 不可以被 client 用来推送消息，仅用来进行 exchange 和 exchange 之间的绑定
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an exchange")

	// 2. 发送消息
	err = r.channel.Publish(
		r.Exchange,
		"", // 订阅模式下 QueueName 就是 ""，由 exchange 自动绑定 queue
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
}

// 订阅模式下消费消息
func (r *RabbitMQPubSub) Consume() {
	// 1. 尝试创建交换机
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an exchange")

	// 2. 尝试创建队列
	q, err := r.channel.QueueDeclare(
		"", // 随机产生队列名称
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
		"", // 在 pub/sub 模式下，key 必须为空
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
	log.Printf("[*] Waiting for messages, To exit press CTRL+C")
	<-forever
}
