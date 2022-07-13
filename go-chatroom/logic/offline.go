package logic

import (
	"container/ring"

	"github.com/spf13/viper"
)

type offlineProcessor struct {
	n int

	// 保存所有用户最近的 n 条信息（环形链表）
	recentRing *ring.Ring

	// 保存某个用户的离线消息（每个用户 n 条）
	userRing map[string]*ring.Ring
}

var OfflineProcessor = newOfflineProcessor()

func newOfflineProcessor() *offlineProcessor {
	n := viper.GetInt("offline-num")

	return &offlineProcessor{
		n:          n,
		recentRing: ring.New(n),
		userRing:   make(map[string]*ring.Ring),
	}
}

func (o *offlineProcessor) Save(msg *Message) {
	if msg.Type != MsgTypeNormal {
		return
	}
	// 将消息存入 recentRing 中并后移一位
	o.recentRing.Value = msg
	o.recentRing = o.recentRing.Next()

	for _, nickname := range msg.Ats {
		nickname = nickname[1:]
		var (
			r  *ring.Ring
			ok bool
		)
		// 懒初始化
		if r, ok = o.userRing[nickname]; !ok {
			r = ring.New(o.n)
		}
		r.Value = msg
		o.userRing[nickname] = r.Next()
	}
}

// 用户离线后再次进入聊天室时，取出相应的消息
func (o *offlineProcessor) Send(user *User) {
	// 取出 n 条聊天记录
	o.recentRing.Do(func(value interface{}) {
		if value != nil {
			user.MessageChannel <- value.(*Message)
		}
	})

	if user.isNew {
		return
	}

	// 取出 n 条 @TA 的消息
	// NOTE: 消息可能重复发送了
	if r, ok := o.userRing[user.NickName]; ok {
		r.Do(func(value interface{}) {
			if value != nil {
				user.MessageChannel <- value.(*Message)
			}
		})

		delete(o.userRing, user.NickName)
	}
}
