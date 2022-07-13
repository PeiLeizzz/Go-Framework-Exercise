package logic

import (
	"go-chatroom/global"
)

/**
 * 单例模式-广播器
 * 要求：
 * 	1. 私有、静态的类实例变量
 * 	2. 构造函数私有化
 * 	3. 静态工厂方法，返回此类的唯一实例
 */
type broadcaster struct {
	// 聊天室的所有用户
	users map[string]*User

	enteringChannel chan *User
	leavingChannel  chan *User
	messageChannel  chan *Message

	// 判断该昵称的用户是否能够进入聊天室
	checkUserChannel      chan string
	checkUserCanInChannel chan bool

	// 获取用户列表
	requestUsersChannel chan struct{}
	usersChannel        chan []*User
}

var Broadcaster = &broadcaster{
	users: make(map[string]*User),

	enteringChannel: make(chan *User),
	leavingChannel:  make(chan *User),
	messageChannel:  make(chan *Message, global.MessageQueueLen),

	checkUserChannel:      make(chan string),
	checkUserCanInChannel: make(chan bool),

	requestUsersChannel: make(chan struct{}),
	usersChannel:        make(chan []*User),
}

func (b *broadcaster) Start() {
	for {
		select {
		case user := <-b.enteringChannel:
			b.users[user.NickName] = user

			OfflineProcessor.Send(user)

		case user := <-b.leavingChannel:
			delete(b.users, user.NickName)
			// 关闭用户接收信息的 goroutine
			user.CloseMessageChannel()

		case msg := <-b.messageChannel:
			for _, user := range b.users {
				if user.UID == msg.User.UID {
					continue
				}
				user.MessageChannel <- msg
			}
			OfflineProcessor.Save(msg)

		case nickname := <-b.checkUserChannel:
			if _, ok := b.users[nickname]; ok {
				b.checkUserCanInChannel <- false
			} else {
				b.checkUserCanInChannel <- true
			}

		case <-b.requestUsersChannel:
			userList := make([]*User, 0, len(b.users))
			for _, user := range b.users {
				userList = append(userList, user)
			}

			b.usersChannel <- userList
		}
	}
}

/**
 * 不使用锁，通过 channel 来通信
 * 并且必须保证 channel 是无缓冲区的
 */
func (b *broadcaster) CanEnterRoom(nickname string) bool {
	b.checkUserChannel <- nickname

	return <-b.checkUserCanInChannel
}

func (b *broadcaster) UserEntering(u *User) {
	b.enteringChannel <- u
}

func (b *broadcaster) UserLeaving(u *User) {
	b.leavingChannel <- u
}

func (b *broadcaster) Broadcast(msg *Message) {
	b.messageChannel <- msg
}

func (b *broadcaster) GetUserList() []*User {
	b.requestUsersChannel <- struct{}{}
	return <-b.usersChannel
}
