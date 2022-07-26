package mq

import (
	"filestore-server/config"
	"github.com/streadway/amqp"
	"log"
)

var (
	conn        *amqp.Connection
	channel     *amqp.Channel
	notifyClose chan *amqp.Error
)

func init() {
	if !config.AsyncTransferEnable {
		return
	}
	if initChannel() {
		channel.NotifyClose(notifyClose) // 异常时发送通知重连
	}

	go func() {
		for {
			select {
			case msg := <-notifyClose:
				conn = nil
				channel = nil
				log.Printf("onNotifyChannelClosed: %+v\n", msg)
				initChannel()
			}
		}
	}()
}

func initChannel() bool {
	// 判断 channel 是否已经创建
	if channel != nil {
		return true
	}

	var err error
	// 获得 rabbitmq 的一个连接
	conn, err = amqp.Dial(config.RabbitURL)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// 打开一个 channel
	channel, err = conn.Channel()
	if err != nil {
		log.Println(err.Error())
		return false
	}

	return true
}

func Publish(exchange string, routingKey string, msg []byte) bool {
	// 判断 channel 是否正常
	if !initChannel() {
		return false
	}

	// 执行消息发布操作
	err := channel.Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		},
	)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	return true
}
