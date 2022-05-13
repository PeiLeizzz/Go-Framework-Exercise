package lru

import "container/list"

type Cache struct {
	maxBytes  int64 // 允许使用的最大内存，设为 0 时表示无限制
	nbytes    int64 // 已使用的内存
	ll        *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value) // 某条记录被移除时的回调函数，可以为 nil
}

// 双向链表中节点的数据类型，保存 key 是为了删除节点时在 map 中也一并删除
type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int // 返回其占用多少 bytes
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

/**
 * 查找功能，包含两个步骤：
 * 1. 从字典中找出对应的双向链表的节点
 * 2. 将该节点移动到队尾
 */
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele) // 这里约定 front 是队尾
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

/**
 * 删除（淘汰），移除队首节点，包含步骤：
 * 1. 取出队首节点
 * 2. 在链表中删除该节点
 * 3. 在 map 中删除该节点对应的映射
 * 4. 更新当前内存
 * 5. 调用回调函数（如果有的话）
 */
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

/**
 * 新增/修改，包含步骤：
 * 1.1 如果键存在，更新对应的 value，并且将其移到队尾
 * 1.2 如果键不存在，在队尾新增节点、并将映射加入 map
 * 2. 更新 nbytes
 * 3. 移除队首节点，直至 c.nbytes <= c.maxBytes
 */
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.nbytes > c.maxBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
