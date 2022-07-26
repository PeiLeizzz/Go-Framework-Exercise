package redis

import (
	"github.com/garyburd/redigo/redis"
	"time"
)

var (
	pool      *redis.Pool
	redisHost = "127.0.0.1:6379"
)

func newRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,
		MaxActive:   30,
		IdleTimeout: 5 * time.Minute,
		Dial: func() (redis.Conn, error) {
			// 打开连接
			c, err := redis.Dial("tcp", redisHost)
			if err != nil {
				return nil, err
			}

			// 访问认证
			//if _, err = c.Do("AUTH", redisPass); err != nil {
			//	c.Close()
			//	return nil, err
			//}
			return c, nil
		},
		// 定时检查健康状况
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

func init() {
	pool = newRedisPool()
}

func RedisPool() *redis.Pool {
	return pool
}
