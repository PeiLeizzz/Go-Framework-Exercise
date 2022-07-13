package server

import (
	"log"
	"net/http"

	"go-chatroom/logic"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

/**
 * 读取连接中的消息的 goroutine
 */
func WebSocketHandleFunc(w http.ResponseWriter, req *http.Request) {
	// 从客户端接收 WebSocket 握手，将连接升级到 WebSocket
	conn, err := websocket.Accept(w, req, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		log.Println("websocket accept error: ", err)
		return
	}

	// 1. 构建新用户实例
	nickname := req.FormValue("nickname")
	token := req.FormValue("token")
	if l := len(nickname); l < 2 || l > 20 {
		log.Println("nickname illegal: ", nickname)
		wsjson.Write(req.Context(), conn, logic.NewErrorMessage("非法昵称，昵称长度：4-20"))
		conn.Close(websocket.StatusUnsupportedData, "nickname illegal!")
		return
	}

	if !logic.Broadcaster.CanEnterRoom(nickname) {
		log.Println("昵称已经存在：", nickname)
		wsjson.Write(req.Context(), conn, logic.NewErrorMessage("该昵称已经存在！"))
		conn.Close(websocket.StatusUnsupportedData, "nickname exists!")
		return
	}

	user := logic.NewUser(conn, nickname, req.RemoteAddr, token)

	// 2. 开启给新用户发送消息的 goroutine
	go user.SendMessage(req.Context())

	// 3. 给新用户发送欢迎信息
	user.MessageChannel <- logic.NewWelcomeMessage(user)

	// 向其他所有用户告知新用户的到来
	msg := logic.NewUserEnterMessage(user)
	logic.Broadcaster.Broadcast(msg)

	// 4. 将该用户加入广播器的用户列表中
	logic.Broadcaster.UserEntering(user)
	log.Println("user: ", nickname, " joins chat")

	// 5. 接收用户消息，该函数当连接断开后才会返回
	err = user.ReceiveMessage(req.Context())

	// 6. 用户离开
	logic.Broadcaster.UserLeaving(user)
	msg = logic.NewUserLeaveMessage(user)
	logic.Broadcaster.Broadcast(msg)
	log.Println("user: ", nickname, " leaves chat")

	// 根据读取时的错误执行不同的 Close
	if err == nil { // 正常断开
		conn.Close(websocket.StatusNormalClosure, "")
	} else {
		log.Println("read from client error: ", err)
		conn.Close(websocket.StatusInternalError, "Read from client error")
	}
}
