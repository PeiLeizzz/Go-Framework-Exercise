package RabbitMQ

import (
	"github.com/streadway/amqp"
	"log"
)

type RabbitTopic struct {
	*rabbitMQ
}

var _ RabbitMQCommon = (*RabbitTopic)(nil)

// Topic 模式创建 RabbitMQ 实例
func NewRabbitMQTopic(exchangeName, routingKey string) *RabbitTopic {
	return &RabbitTopic{
		newRabbitMQ("", exchangeName, routingKey),
	}
}

// Topic 模式发送消息
func (r *RabbitTopic) Publish(message string) {
	// 1. 尝试创建交换机
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		"topic", // 注意这里是 topic,
		true,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare on exchange")

	// 2. 发送消息
	err = r.channel.Publish(
		r.Exchange,
		r.Key,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
}

// Topic 模式消费消息
// 其中 * 用于匹配一个单词
// # 用于匹配多个单词（可以是 0 个）
func (r *RabbitTopic) Consume() {
	// 1. 尝试创建交换机
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declared an exchange")

	// 2. 尝试创建队列，这里队列名也省略
	q, err := r.channel.QueueDeclare(
		"",
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
	log.Printf("[*] Waiting for messages, To exit press CTRL+C")
	<-forever
}