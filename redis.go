package main

import (
	"log"

	"github.com/gomodule/redigo/redis"
)

// RedisHub is ...
type RedisHub struct {
	pool     *redis.Pool
	channels map[string]*redis.PubSubConn // map to hold red.Pubsub types
	// RedisHub has subclient goroutine method to create a new pub sub connnection to specified channel
}

func createRedisHub(host string) *RedisHub {
	pool := &redis.Pool{
		IdleTimeout: 0,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", host)
		},
	}

	return &RedisHub{
		pool:     pool,
		channels: make(map[string]*redis.PubSubConn),
	}
}

func (r *RedisHub) subClient(channel string, ch chan *Message) {
	newPool := r.pool.Get()
	psc := redis.PubSubConn{Conn: newPool}
	psc.Subscribe(channel)
	r.channels[channel] = &psc

	for {
		defer func() {
			psc.Close()
			newPool.Close()
		}()
		switch v := psc.Receive().(type) {
		case redis.Message:
			ch <- &Message{v.Channel, string(v.Data)}
		case redis.Subscription:
			// log.Printf("Subscribed to redis pub sub channel %s: %s %d\n", v.Channel, v.Kind, v.Count)
			// https://godoc.org/github.com/garyburd/redigo/redis#Subscription
			// need to check if it is type unsubscribe and end goroutine
			log.Println(v)
			if v.Kind == "unsubscribe" || v.Kind == "punsubscribe" {
				delete(r.channels, channel)
				psc.Close()
				return
			}
		case error:
			log.Printf("redis pubsub receive err: %v\n", v)
			psc.Close()
			panic("Redis Sub connection broke")
		default:
			log.Println("something else happened")
		}
	}
}
