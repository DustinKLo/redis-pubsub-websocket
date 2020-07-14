package main

import (
	"log"

	"github.com/gomodule/redigo/redis"
)

// RedisHub is ...
type RedisHub struct {
	psc         *redis.PubSubConn
	subscribe   chan string // send to central redis psc to sub or unsub to channel
	unsubscribe chan string
}

func newRedisPool(host string) *redis.Pool {
	pool := &redis.Pool{
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
	return pool
}

func newRedisHub(conn *redis.Conn) *RedisHub {
	return &RedisHub{
		psc: &redis.PubSubConn{
			Conn: *conn,
		},
		subscribe:   make(chan string),
		unsubscribe: make(chan string),
	}
}

func (r *RedisHub) subscribeHandler() {
	for {
		select {
		case channel := <-r.subscribe:
			r.psc.Subscribe(channel)
		case channel := <-r.unsubscribe:
			r.psc.Unsubscribe(channel)
		}
	}
}

func (r *RedisHub) subClient(ch chan *Message) {
	for {
		defer r.psc.Close()
		switch v := r.psc.Receive().(type) {
		case redis.Message:
			ch <- &Message{v.Channel, v.Data}
		case redis.Subscription:
			// https://godoc.org/github.com/garyburd/redigo/redis#Subscription
			log.Println(v)
		case error:
			log.Printf("redis pubsub receive err: %v\n", v)
			panic("Redis Sub connection broke")
		default:
			log.Println("something else happened")
		}
	}
}
