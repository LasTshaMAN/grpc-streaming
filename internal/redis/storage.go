package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/LasTshaMAN/streaming"
)

// TODO
// Pass in a logger here to log connection issues (when closing) and unexpected Redis replies

type Storage struct {
	pool *redis.Pool
}

func NewStorage(pool *redis.Pool) *Storage {
	return &Storage{
		pool: pool,
	}
}

func (storage *Storage) Get(ctx context.Context, url string) (string, time.Duration, error) {
	conn, err := storage.pool.GetContext(ctx)
	if err != nil {
		return "", 0, fmt.Errorf("get context, err: %w", err)
	}
	defer conn.Close()

	// TODO
	// Combine value and ttl fetch into a single command for better performance.

	value, err := redis.String(conn.Do("GET", url))
	if err != nil {
		if err == redis.ErrNil {
			return "", 0, streaming.ErrDataNotFoundInStorage
		}
		return "", 0, fmt.Errorf("get value from Redis, err: %w", err)
	}

	ttlSeconds, err := redis.Int(conn.Do("TTL", url))
	if err != nil {
		if err == redis.ErrNil {
			return value, 0, nil
		}
		return "", 0, fmt.Errorf("get ttl from Redis, err: %w", err)
	}

	return value, time.Duration(ttlSeconds) * time.Second, nil
}

func (storage *Storage) Set(ctx context.Context, url string, data string, ttl time.Duration) error {
	conn, err := storage.pool.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("get context, err: %w", err)
	}
	defer conn.Close()

	_, err = conn.Do("SETEX", url, int(ttl/time.Second), data)
	if err != nil {
		return fmt.Errorf("set value with ttl in Redis, err: %w", err)
	}

	return nil
}
