package RabbitMQ

import (
	"log"

	"github.com/streadway/amqp"
)

// amqp://账号:密码@地址:端口号/vhost
const MQURL = "amqp://peileiuser:peilei777@127.0.0.1:5672/peilei"

type RabbitMQCommon interface {
	Publish(string)
	Consume()
}

type rabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	// 队列名称
	QueueName string
	// 交换机
	Exchange string
	// routingkey
	Key string
	// 连接信息
	Mqurl string
}

// 创建 RabbitMQ 结构体实例
func newRabbitMQ(queueName, exchange, key string) *rabbitMQ {
	rabbitmq := rabbitMQ{
		QueueName: queueName,
		Exchange:  exchange,
		Key:       key,
		Mqurl:     MQURL,
	}

	var err error
	// 创建 rabbitmq 连接
	rabbitmq.conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "创建连接错误")
	// 获取 channel
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "获取 channel 失败")
	return &rabbitmq
}

// 错误处理函数
func (r *rabbitMQ) failOnErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s:%s", message, err)
	}
}

// 断开连接
func (r *rabbitMQ) Destroy() {
	r.channel.Close()
	r.conn.Close()
}
