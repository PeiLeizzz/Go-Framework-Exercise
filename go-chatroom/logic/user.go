package logic

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"go-chatroom/global"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type User struct {
	UID            int           `json:"uid"`
	NickName       string        `json:"nickname"`
	EnterAt        time.Time     `json:"enter_at"`
	Addr           string        `json:"addr"`
	MessageChannel chan *Message `json:"-"`
	Token          string        `json:"token"`

	isNew bool

	conn *websocket.Conn
}

var System = &User{}
var globalUID uint32 = 0

func NewUser(conn *websocket.Conn, nickname, addr, token string) *User {
	user := &User{
		NickName:       nickname,
		Addr:           addr,
		EnterAt:        time.Now(),
		MessageChannel: make(chan *Message, global.MessageQueueLen),
		Token:          token,

		conn: conn,
	}

	if user.Token != "" {
		uid, err := parseTokenAndValidate(token, nickname)
		if err == nil {
			user.UID = uid
		}
	}
	if user.UID == 0 {
		user.UID = int(atomic.AddUint32(&globalUID, 1))
		user.Token = genToken(user.UID, user.NickName)
		user.isNew = true
	}

	return user
}

func (u *User) SendMessage(ctx context.Context) {
	for msg := range u.MessageChannel {
		wsjson.Write(ctx, u.conn, msg)
	}
}

func (u *User) ReceiveMessage(ctx context.Context) error {
	var (
		receiveMsg map[string]string
		err        error
	)

	for {
		err = wsjson.Read(ctx, u.conn, &receiveMsg)
		if err != nil {
			var closeErr websocket.CloseError
			// 判断是否是正常关闭导致的 err
			if errors.As(err, &closeErr) {
				return nil
			}

			return err
		}

		sendMsg := NewMessage(u, FilterSensitive(receiveMsg["content"]), receiveMsg["send_time"])

		reg := regexp.MustCompile(`@[^\s@]{2,20}`) // 昵称在 2～20 字符之间
		sendMsg.Ats = reg.FindAllString(sendMsg.Content, -1)

		Broadcaster.Broadcast(sendMsg)
	}
}

func FilterSensitive(content string) string {
	for _, word := range global.SensitiveWords {
		content = strings.ReplaceAll(content, word, "**")
	}

	return content
}

func (u *User) CloseMessageChannel() {
	close(u.MessageChannel)
}

/**
 * 生成 token 的步骤：
 * 	1. message: nickname + secret + uid
 * 	2. 对 message 使用 HMAC-SHA256 算法计算出 Hash 值，记为 messageMAC
 * 	3. 对 messageMAC 进行 Base64 处理，得到 messageMACStr
 * 	4. token: messageMACStr + "uid" + uid
 */
func genToken(uid int, nickname string) string {
	secret := viper.GetString("token-secret")
	message := fmt.Sprintf("%s%s%d", nickname, secret, uid)

	messageMAC := macSha256([]byte(message), []byte(secret))

	return fmt.Sprintf("%suid%d", base64.StdEncoding.EncodeToString(messageMAC), uid)
}

func macSha256(message, secret []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write(message)
	return mac.Sum(nil)
}

func parseTokenAndValidate(token, nickname string) (int, error) {
	pos := strings.LastIndex(token, "uid")
	messageMAC, err := base64.StdEncoding.DecodeString(token[:pos])
	if err != nil {
		return 0, err
	}
	uid := cast.ToInt(token[pos+3:])

	secret := viper.GetString("token-secret")
	message := fmt.Sprintf("%s%s%d", nickname, secret, uid)
	ok := validateMAC([]byte(message), []byte(secret), messageMAC)
	if ok {
		return uid, nil
	}

	return 0, errors.New("token is illegal")
}

func validateMAC(message, secret, messageMAC []byte) bool {
	expectedMAC := macSha256(message, secret)
	return hmac.Equal(messageMAC, expectedMAC)
}
