package lfu_test

import (
	"github.com/matryer/is"
	"go-cache/lfu"
	"testing"
)

func TestSetGet(t *testing.T) {
	is := is.New(t)

	cache := lfu.New(24, nil)
	cache.DelOldest()
	cache.Set("k1", 1)
	v := cache.Get("k1")
	is.Equal(v, 1)

	cache.Del("k1")
	is.Equal(0, cache.Len())
}

func TestOnEvicted(t *testing.T) {
	is := is.New(t)

	keys := make([]string, 0, 8)
	onEvicted := func(key string, value interface{}) {
		keys = append(keys, key)
	}

	cache := lfu.New(32, onEvicted)

	cache.Set("k1", 1)
	cache.Set("k2", 2)
	cache.Get("k1")
	cache.Get("k1")
	cache.Get("k2")
	cache.Set("k3", 3) // k3 out
	cache.Set("k4", 4) // k4 out

	expected := []string{"k3", "k4"}

	is.Equal(expected, keys)
	is.Equal(2, cache.Len())
}
