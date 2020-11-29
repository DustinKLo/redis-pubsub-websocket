package main

import (
	"log"

	"github.com/gomodule/redigo/redis"
)

func newRedisClient(host string) *redis.Pool {
	redisPool := &redis.Pool{
		IdleTimeout: 0,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.DialURL(host)
			if err != nil {
				log.Printf(err.Error())
				panic("ERROR: failed to initialize Redis Pool")
			}
			return conn, err
		},
	}
	return redisPool
}

func newPubsubClient(pool *redis.Pool) *redis.PubSubConn {
	return &redis.PubSubConn{
		Conn: pool.Get(),
	}
}
