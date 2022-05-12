/**
 * 缓存值的抽象与封装
 */
package geecache

/**
 * 一个只读数据结构，用来表示缓存值
 */
type ByteView struct {
	b []byte
}

func (v ByteView) Len() int {
	return len(v.b)
}

// 返回 v.b 的副本
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
