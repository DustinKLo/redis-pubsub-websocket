package main

import (
	"log"
	"sync"

	"github.com/gomodule/redigo/redis"
)

// RedisHub is ...
type RedisHub struct {
	pool     *redis.Pool
	channels map[string]*redis.PubSubConn // map to hold red.Pubsub types
	// RedisHub has subclient goroutine method to create a new pub sub connnection to specified channel
	mtx sync.Mutex
}

func newRedisHub(host string) *RedisHub {
	pool := &redis.Pool{
		IdleTimeout: 0,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", host)
			if err != nil {
				log.Printf("ERROR: fail initializing the redis pool: %s", err.Error())
				panic("ERROR: failed to initialize Redis Pool")
			}
			return conn, err
		},
	}

	return &RedisHub{
		pool:     pool,
		channels: make(map[string]*redis.PubSubConn),
		mtx:      sync.Mutex{},
	}
}

func (r *RedisHub) subClient(channel string, ch chan *Message) {
	pool := r.pool.Get()
	psc := redis.PubSubConn{Conn: pool}
	psc.Subscribe(channel)

	r.mtx.Lock()
	r.channels[channel] = &psc
	r.mtx.Unlock()

	for {
		defer func() {
			psc.Close()
			pool.Close()
		}()
		switch v := psc.Receive().(type) {
		case redis.Message:
			log.Println(string(v.Data))
			ch <- &Message{v.Channel, string(v.Data)}
		case redis.Subscription:
			// log.Printf("Subscribed to redis pub sub channel %s: %s %d\n", v.Channel, v.Kind, v.Count)
			// https://godoc.org/github.com/garyburd/redigo/redis#Subscription
			// need to check if it is type unsubscribe and end goroutine
			log.Println(v)
			if v.Kind == "unsubscribe" || v.Kind == "punsubscribe" {
				r.mtx.Lock()
				delete(r.channels, channel)
				r.mtx.Unlock()
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
