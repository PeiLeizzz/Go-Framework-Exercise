package mq

import "log"

var done chan struct{}

func StartConsume(qName string, cName string, callback func(msg []byte) bool) {
	// 通过 channel.Consume 获得消息信道
	msgs, err := channel.Consume(
		qName,
		cName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Println(err.Error())
		return
	}

	done = make(chan struct{})

	// 从信道中循环获取消息
	go func() {
		for msg := range msgs {
			// 获取到消息后用 callback 处理
			ok := callback(msg.Body)
			if !ok {
				// TODO: 将任务写到另一个队列，用于异常情况的重试
			}
		}
	}()

	<-done

	channel.Close()
}

func StopConsume() {
	done <- struct{}{}
}
