package main

import (
	"log"

	"github.com/gomodule/redigo/redis"
)

func newRedisPool(host string) *redis.Pool {
	return &redis.Pool{
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
}

func newPubsubClient(pool *redis.Pool) *redis.PubSubConn {
	return &redis.PubSubConn{
		Conn: pool.Get(),
	}
}
