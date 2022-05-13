package singleflight

import "sync"

/**
 * 代表正在进行中，或已经结束的请求
 */
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

/**
 * 管理不同 key 的请求
 */
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

/**
 * 保证针对相同的 key，并发条件下无论 Do 被[瞬时]调用多少次，fn 都只会被调用一次
 */
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	// 延迟初始化
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	// 如果有正在进行中的 key 对应的请求，则等待其完成
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
