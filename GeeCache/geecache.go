/**
 * 负责与外部交互，控制缓存存储和获取的主流程
 */
package geecache

import (
	"fmt"
	"log"
	"sync"
)

/**
 * 当缓存不存在时，调用的回调函数（接口），主要是用于获取源数据
 * 具体如何从源获取数据，由用户决定即可
 */
type Getter interface {
	Get(key string) ([]byte, error)
}

/**
 * 接口型函数，方便使用者在使用接口参数时，既能传入函数作为参数
 * 也能够传入实现了该接口的结构体作为参数
 */
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

/**
 * 负责与用户交互，控制缓存值存储和获取
 * 一个 Group 可以认为是一个缓存的命名空间（例如缓存成绩的 Group 命名为 scores）
 */
type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	mu.Lock()
	defer mu.Unlock()
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	// 缓存未命中，从数据源中加载数据
	return g.load(key)
}

/**
 * 从数据源中加载数据
 * 分布式场景下会询问其他节点
 */
func (g *Group) load(key string) (ByteView, error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	// 通过回调函数加载数据
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	// 获取数据的副本，主要是为了保存一个副本
	// 防止此时 getter.Get() 后，外部仍然掌握有 bytes 的修改权
	// 导致保存后，切片被外部修改
	value := ByteView{b: cloneBytes(bytes)}
	// 加入缓存
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
