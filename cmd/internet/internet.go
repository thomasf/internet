package main

import (
	"flag"
	"time"

	"github.com/garyburd/redigo/redis"
)

func main() {
	flag.Parse()

	pool := newPool(*redisServer, *redisPassword)

}

func newPool(server, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

}

var (
	pool          *redis.Pool
	redisServer   = flag.String("redisServer", ":28743", "")
	redisPassword = flag.String("redisPassword", "", "")
	ip2asn        = flag.String("ip2asn", "", "IP address to resolve to AS number")
)
