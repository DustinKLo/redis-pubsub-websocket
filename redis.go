package main

import (
	"log"

	"github.com/gomodule/redigo/redis"
)

func redisConn(host string) redis.Conn {
	c, err := redis.Dial("tcp", host)
	if err != nil {
		panic(err)
	}
	return c
}

func subClient(psc redis.PubSubConn, ch chan *Message) {
	for {
		defer psc.Close()
		switch v := psc.Receive().(type) {
		case redis.Message:
			ch <- &Message{v.Channel, string(v.Data)}
		case redis.Subscription:
			// log.Printf("Subscribed to redis pub sub channel %s: %s %d\n", v.Channel, v.Kind, v.Count)
			// https://godoc.org/github.com/garyburd/redigo/redis#Subscription
			// need to check if it is type unsubscribe
			// return Nil if unsubscribe or punsubscribe
		case error:
			log.Printf("redis pubsub receive err: %v\n", v)
			panic("Redis Sub connection broke")
		default:
			log.Println("something else happened")
		}
	}
}
