package common

import (
	"errors"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

type uints []uint32

func (x uints) Len() int {
	return len(x)
}

func (x uints) Less(i, j int) bool {
	return x[i] < x[j]
}

func (x uints) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

var errEmpty = errors.New("Hash 环没有数据")

const VIRTUAL_NODE_NUM = 20

type Consistent struct {
	circle       map[uint32]string // 节点的哈希值: 节点的 key
	sortedHashes uints             // 已经排序过的节点切片
	VirtualNode  int               // 虚拟节点倍数
	mu           sync.RWMutex
}

func NewConsistent() *Consistent {
	return &Consistent{
		circle:      make(map[uint32]string),
		VirtualNode: VIRTUAL_NODE_NUM,
	}
}

// 自动生成 key 值, element 可以是 ip 地址等节点的唯一信息
func (c *Consistent) generateKey(element string, index int) string {
	return element + strconv.Itoa(index)
}

func (c *Consistent) hashKey(key string) uint32 {
	if len(key) < 64 {
		dst := make([]byte, 64)
		copy(dst, key)
		return crc32.ChecksumIEEE(dst[:len(key)])
	}
	return crc32.ChecksumIEEE([]byte(key))
}

func (c *Consistent) updateSortedHashes() {
	hashes := c.sortedHashes[:0]
	// 判断切片容量，是否过大，如果过大则重置切片（有效节点不足 1/4 时）
	// (c.VirtualNode * capNode) / (c.VirtualNode * 4) > (c.VirtualNode * validNode)
	// -> capNode / 4 > validNode
	if cap(c.sortedHashes)/(c.VirtualNode*4) > len(c.circle) {
		hashes = nil
	}

	// 重新排序
	for k := range c.circle {
		hashes = append(hashes, k)
	}
	sort.Sort(hashes)
	c.sortedHashes = hashes
}

func (c *Consistent) add(element string) {
	for i := 0; i < c.VirtualNode; i++ {
		c.circle[c.hashKey(c.generateKey(element, i))] = element
	}

	// 更新排序
	c.updateSortedHashes()
}

// 增加节点
func (c *Consistent) Add(element string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.add(element)
}

func (c *Consistent) remove(element string) {
	for i := 0; i < c.VirtualNode; i++ {
		delete(c.circle, c.hashKey(c.generateKey(element, i)))
	}

	c.updateSortedHashes()
}

// 删除节点
func (c *Consistent) Remove(element string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.remove(element)
}

func (c *Consistent) search(key uint32) int {
	f := func(x int) bool {
		return c.sortedHashes[x] > key
	}
	// 二分查找
	i := sort.Search(len(c.sortedHashes), f)

	if i >= len(c.sortedHashes) {
		i = 0
	}
	return i
}

// 根据数据获取最近的服务器信息
func (c *Consistent) Get(name string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.circle) == 0 {
		return "", errEmpty
	}

	key := c.hashKey(name)
	// 查找环上最近的节点（顺时针）
	i := c.search(key)
	// c.sortedHashes[i] 保存的是服务器节点的哈希值
	return c.circle[c.sortedHashes[i]], nil
}
