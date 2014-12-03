package main

import (
	"github.com/garyburd/redigo/redis"
	"time"
)

func initRedis(host string) *redis.Pool {
	return &redis.Pool{
		MaxIdle: 64,
		//MaxActive:   10240,
		IdleTimeout: 60 * time.Second,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")

			return err
		},
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host)
			if err != nil {
				return nil, err
			}

			return c, err
		},
	}
}

func CacheGet(rd redis.Conn, key string) string {
	s, err := redis.String(rd.Do("GET", key))
	if err != nil {
		pp(err.Error())
	}
	if s == "null" {
		return ""
	}
	return s
}

func CacheSet(rd redis.Conn, key string, val string, time int) {
	_, err := rd.Do("SET", key, val)
	if err != nil {
		pp(err.Error())
	}

	rd.Do("EXPIRE", key, time)
}

func RedisExec(rd redis.Conn, cmd string, args ...interface{}) (interface{}, error) {
	return rd.Do(cmd, args...)
}

func RedisSend(rd redis.Conn, cmd string, args ...interface{}) error {
	return rd.Send(cmd, args...)
}
