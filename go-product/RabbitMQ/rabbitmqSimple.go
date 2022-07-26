package RabbitMQ

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"go-product/imooc-product/datamodels"
	"go-product/imooc-product/services"
	"log"
	"sync"
)

type RabbitMQSimple struct {
	*rabbitMQ
	mu sync.Mutex
}

var _ RabbitMQCommon = (*RabbitMQSimple)(nil)

// Simple 模式下创建 RabbitMQ 实例
func NewRabbitMQSimple(queueName string) *RabbitMQSimple {
	// Simple 模式下 exchange 和 key 默认为空
	// 即代表使用 default exchange(名字为 "" 的 direct exchange)
	// 这个 exchange 默认绑定到与 queue 同名的 routingkey 上
	return &RabbitMQSimple{
		newRabbitMQ(queueName, "", ""),
		sync.Mutex{},
	}
}

// Simple 模式下生产消息
func (r *RabbitMQSimple) Publish(message string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// 1. 申请队列（如果队列不存在，则创建）
	// 	  保证消息能够发送到队列中
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		false, // 控制消息是否持久化
		false, // 最后一个消费者断开连接后，是否将队列删除
		false, // 是否具有排他性，如果设置为 true，那就仅当前用户可见
		false, // 是否不阻塞（设为 false 则要等待服务器的响应）
		nil,   // 额外属性
	)
	if err != nil {
		fmt.Println(err)
	}
	// 2. 发送消息到队列中
	r.channel.Publish(
		r.Exchange,  // 在 Simple 模式下，其实就是 ""
		r.QueueName, // routingkey 一般就用 QueueName
		false,       // 如果为 true，会根据 exchange 类型和 routingkey 规则，判断是否有符合条件的队列，如果没有则会把发送的消息回退给发送者
		false,       // 如果为 true，当 exchange 发送消息到队列后发现队列上没有绑定消费者，则会把消息返还给发送者
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
}

// Simple 模式下消费消息
func (r *RabbitMQSimple) Consume() {
	// 1. 申请队列（如果队列不存在，则创建）
	// 	  保证消息能够发送到队列中
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		false, // 控制消息是否持久化
		false, // 最后一个消费者断开连接后，是否将队列删除
		false, // 是否具有排他性，如果设置为 true，那就仅当前用户可见
		false, // 是否不阻塞（设为 false 则要等待服务器的响应）
		nil,   // 额外属性
	)
	if err != nil {
		fmt.Println(err)
	}

	// 2. 接收消息
	msgs, err := r.channel.Consume(
		r.QueueName, // routingkey 这里和生产者一致
		"",          // 用来区分多个消费者
		true,        // 是否自动应答（告诉服务器是否消费完）
		false,       // 是否具有排他性
		false,       // 如果设置为 true，表示不能将同一个 connection 中发送的消息传递给同一个 connection 中的消费者
		false,       // 队列消费是否不阻塞
		nil,
	)

	if err != nil {
		fmt.Println(err)
	}

	// 3. 消费消息
	forever := make(chan bool)
	go func() {
		for d := range msgs {
			// 实现我们要处理的逻辑函数
			log.Printf("Received a message: %s", d.Body)
		}
	}()
	log.Printf("[*] Waiting for messages, To exit press CTRL+C")
	<-forever
}

func (r *RabbitMQSimple) ConsumeOrder(orderService services.IOrderService, productService services.IProductService) {
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}

	// 2. 接收消息
	msgs, err := r.channel.Consume(
		r.QueueName,
		"",
		false, // 手动应答
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		fmt.Println(err)
	}

	// 消息流控，防止爆库
	r.channel.Qos(
		1,     // 当前消费者一次能接收的最大消息数量
		0,     // 服务器传递的最大容量（以 8 位字节为单位）
		false, // 如果为 true，表示对整个 channel 可用
	)

	// 3. 消费消息
	forever := make(chan bool)
	go func() {
		for d := range msgs {
			// 实现我们要处理的逻辑函数
			log.Printf("Received a message: %s", d.Body)

			message := &datamodels.Message{}
			err := json.Unmarshal([]byte(d.Body), message)
			if err != nil {
				fmt.Println(err)
			}

			_, err = orderService.InsertOrderByMessage(message)
			if err != nil {
				fmt.Println(err)
			}

			err = productService.SubNumberOne(message.ProductID)
			if err != nil {
				fmt.Println(err)
			}
			// 手动应答，让服务器删除消息
			// 如果为 true，表示确认所有未确认的消息
			// 如果为 false，表示确认当前消息
			d.Ack(false)
		}
	}()
	log.Printf("[*] Waiting for messages, To exit press CTRL+C")
	<-forever
}
