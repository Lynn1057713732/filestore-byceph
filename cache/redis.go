package redis

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"time"
)

var (
	pool *redis.Pool
	redisHost = "127.0.0.1:6379"
	redisPass = "1057713732"
)

func NewRedisPool() *redis.Pool{
	return &redis.Pool{
		MaxIdle: 50,
		MaxActive: 30,
		IdleTimeout: time.Second * 300,
		Dial: func() (redis.Conn, error) {
			// 1.打开链接
			c, err := redis.Dial("tcp", redisHost)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}

			//2.访问认证
			if _, err := c.Do("AUTH", redisPass);err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := conn.Do("PING")
			return err
		},
	}
}

func init() {
	pool = NewRedisPool()
}

func RedisPool() *redis.Pool{
	return pool
}

