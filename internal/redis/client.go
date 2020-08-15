package redis

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

// NewClient returns new client.
func NewClient(
	host string,
	db int,
	dialTimeout time.Duration,
	readTimeout time.Duration,
	writeTimeout time.Duration,
	maxIdle int,
	maxActive int,
	idleTimeout time.Duration,
) *redis.Pool {
	pool := &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: idleTimeout,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(
				"tcp",
				host,
				redis.DialConnectTimeout(dialTimeout),
				redis.DialReadTimeout(readTimeout),
				redis.DialWriteTimeout(writeTimeout),
			)
			if err != nil {
				return nil, fmt.Errorf("dial redis: %w", err)
			}
			if _, err = c.Do("SELECT", db); err != nil {
				c.Close()
				return nil, fmt.Errorf("select redis db: %w", err)
			}
			return c, nil
		},
		// Check the health of an idle connection before the connection is returned
		// to the application. It PINGs connections that have been idle more than a minute.
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				return fmt.Errorf("ping redis: %w", err)
			}
			return nil
		},
	}

	return pool
}
