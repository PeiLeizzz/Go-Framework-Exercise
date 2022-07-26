package fifo

import (
	"container/list"
	"github.com/PeiLeizzz/cache"
)

type fifo struct {
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
	return &fifo{
		maxBytes:  maxBytes,
		onEvicted: onEvicted,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
	}
}

func (f *fifo) Set(key string, value interface{}) {
	// 如果 key 已存在，则更新 value，并将其移到队尾
	if e, ok := f.cache[key]; ok {
		f.ll.MoveToBack(e)
		en := e.Value.(*entry)
		// 更新 value 占用的内存
		f.usedBytes = f.usedBytes - en.Len() + cache.CalcLen(value)
		en.value = value
		return
	}

	en := &entry{key, value}
	e := f.ll.PushBack(en)
	f.cache[key] = e

	f.usedBytes += en.Len()
	for f.maxBytes > 0 && f.usedBytes > f.maxBytes {
		f.DelOldest()
	}
}

func (f *fifo) Get(key string) interface{} {
	if e, ok := f.cache[key]; ok {
		return e.Value.(*entry).value
	}

	return nil
}

func (f *fifo) Del(key string) {
	if e, ok := f.cache[key]; ok {
		f.removeElement(e)
	}
}

func (f *fifo) DelOldest() {
	f.removeElement(f.ll.Front())
}

func (f *fifo) removeElement(e *list.Element) {
	if e == nil {
		return
	}

	f.ll.Remove(e)

	en := e.Value.(*entry)
	f.usedBytes -= en.Len()

	delete(f.cache, en.key)

	if f.onEvicted != nil {
		f.onEvicted(en.key, en.value)
	}
}

// 返回记录数
func (f *fifo) Len() int {
	return f.ll.Len()
}
