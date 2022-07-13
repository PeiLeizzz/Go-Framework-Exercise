package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", ":2020")
	if err != nil {
		panic(err)
	}

	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleConn(conn)
	}
}

var (
	// 负责用户加入时 对存储用户信息 map 的通信
	enteringChannel = make(chan *User)
	// 负责用户离开时 对存储用户信息 map 的通信
	leavingChannel = make(chan *User)
	// 负责用户普通消息的广播，缓冲用于避免用户发消息的阻塞
	messageChannel = make(chan *Message, 8)
)

type User struct {
	ID             int
	Addr           string
	EnterAt        time.Time
	MessageChannel chan *Message
}

func (u *User) String() string {
	return "ip: " + u.Addr + ", uid: " + strconv.Itoa(u.ID) + ", enter at: " + u.EnterAt.Format("2006-01-02 15:04:05")
}

var (
	mu    sync.Mutex
	curID = 0
)

func genUserID() int {
	mu.Lock()
	defer mu.Unlock()
	curID++
	return curID
}

func getTimeString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

type Message struct {
	OwnerID int
	Content string
}

/**
 * 处理与客户端的连接
 * tips:
 * 	1. 用户加入时，先向其他人广播，再存储该用户
 * 	2. 用户离开时，先删除该用户，再向其他人广播
 */
func handleConn(conn net.Conn) {
	defer conn.Close()

	user := &User{
		ID:             genUserID(),
		Addr:           conn.RemoteAddr().String(),
		EnterAt:        time.Now(),
		MessageChannel: make(chan *Message, 8),
	}

	// 写 goroutinue（user 接收信息）
	go sendMessage(conn, user.MessageChannel)

	msg := &Message{
		OwnerID: user.ID,
	}
	// 向 user 发欢迎信息
	msg.Content = "Welcome, " + user.String()
	user.MessageChannel <- msg

	// 向其他人广播消息
	msg.Content = getTimeString() + "-user `" + strconv.Itoa(user.ID) + "` has entered"
	messageChannel <- msg

	// 记录到全局用户列表中，避免用锁
	enteringChannel <- user

	// 检查活跃程度
	userActive := make(chan struct{})
	go checkActive(conn, userActive)

	// 用户发送消息
	input := bufio.NewScanner(conn)
	for input.Scan() {
		msg.Content = getTimeString() + "-user " + strconv.Itoa(user.ID) + ": " + input.Text()
		messageChannel <- msg

		// 发消息代表其活跃
		userActive <- struct{}{}
	}
	if err := input.Err(); err != nil {
		log.Println("读取错误：", err)
	}

	// 用户离开
	leavingChannel <- user
	msg.Content = getTimeString() + "-user `" + strconv.Itoa(user.ID) + "` has left"
	messageChannel <- msg
}

/**
 * 通过定时器检查用户活跃度
 * 如果长时间不活跃则关闭其连接
 */
func checkActive(conn net.Conn, userActive <-chan struct{}) {
	d := 10 * time.Second
	timer := time.NewTimer(d)
	for {
		select {
		case <-timer.C:
			conn.Close()
			return
		case <-userActive:
			timer.Reset(d)
		}
	}
}

/**
 * 向 conn 连接发送消息
 * tips: 要注意 ch 的关闭
 */
func sendMessage(conn net.Conn, ch <-chan *Message) {
	for msg := range ch {
		fmt.Fprintln(conn, msg.Content)
	}
}

/**
 * 广播消息
 * 1. 新用户加入
 * 2. 用户普通信息
 * 3. 用户离开
 */
func broadcaster() {
	users := make(map[*User]struct{})

	for {
		select {
		case user := <-enteringChannel:
			users[user] = struct{}{}
		case user := <-leavingChannel:
			delete(users, user)
			// 避免 goroutine 泄露
			close(user.MessageChannel)
		case msg := <-messageChannel:
			for user := range users {
				if user.ID == msg.OwnerID {
					continue
				}
				user.MessageChannel <- msg
			}
		}
	}
}
