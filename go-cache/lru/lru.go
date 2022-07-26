package lru

import (
	"container/list"
	"github.com/PeiLeizzz/cache"
)

type lru struct {
	// 缓存的最大容量（字节）
	maxBytes int
	// 移除一个数据时的回调函数，key 应是可比较类型
	onEvicted func(key string, value interface{})
	// 已使用的字节数
	usedBytes int

	ll *list.List
	// key: value
	cache map[string]*list.Element
}

// Element 中存放的值（指针）
type entry struct {
	// 保留 key 是为了删除时使用
	key   string
	value interface{}
}

func (e *entry) Len() int {
	return cache.CalcLen(e.value)
}

func New(maxBytes int, onEvicted func(key string, value interface{})) cache.Cache {
	return &lru{
		maxBytes:  maxBytes,
		onEvicted: onEvicted,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
	}
}

func (l *lru) Set(key string, value interface{}) {
	// 如果 key 已存在，则更新 value，并将其移到队尾
	if e, ok := l.cache[key]; ok {
		l.ll.MoveToBack(e)
		en := e.Value.(*entry)
		// 更新 value 占用的内存
		l.usedBytes = l.usedBytes - en.Len() + cache.CalcLen(value)
		en.value = value
		return
	}

	en := &entry{key, value}
	e := l.ll.PushBack(en)
	l.cache[key] = e

	l.usedBytes += en.Len()
	for l.maxBytes > 0 && l.usedBytes > l.maxBytes {
		l.DelOldest()
	}
}

func (l *lru) Get(key string) interface{} {
	if e, ok := l.cache[key]; ok {
		l.ll.MoveToBack(e) // 与 FIFO 的唯一不同
		return e.Value.(*entry).value
	}

	return nil
}

func (l *lru) Del(key string) {
	if e, ok := l.cache[key]; ok {
		l.removeElement(e)
	}
}

func (l *lru) DelOldest() {
	l.removeElement(l.ll.Front())
}

func (l *lru) removeElement(e *list.Element) {
	if e == nil {
		return
	}

	l.ll.Remove(e)

	en := e.Value.(*entry)
	l.usedBytes -= en.Len()

	delete(l.cache, en.key)

	if l.onEvicted != nil {
		l.onEvicted(en.key, en.value)
	}
}

// 返回记录数
func (l *lru) Len() int {
	return l.ll.Len()
}
