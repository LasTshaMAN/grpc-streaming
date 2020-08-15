package redis_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/LasTshaMAN/streaming"
	"github.com/LasTshaMAN/streaming/internal/redis"
)

func TestStorage(t *testing.T) {
	const (
		host = "localhost:6379"
		db   = 0

		url  = "some url"
		data = "some data"
		ttl  = 2 * time.Second
	)

	ctx := context.Background()

	t.Run("get existent", func(t *testing.T) {
		client := redis.NewClient(host, db, time.Minute, time.Minute, time.Minute, 16, 16, time.Minute)
		defer func() {
			err := client.Close()

			assert.Nil(t, err)
		}()

		flushRedis(t, client)

		storage := redis.NewStorage(client)

		err := storage.Set(ctx, url, data, ttl)

		assert.Nil(t, err)

		actData, actTTL, err := storage.Get(ctx, url)

		assert.Nil(t, err)
		assert.Equal(t, data, actData)
		assert.True(t, actTTL > 0)
		assert.True(t, actTTL <= ttl)
	})
	t.Run("get non-existent", func(t *testing.T) {
		client := redis.NewClient(host, db, time.Minute, time.Minute, time.Minute, 16, 16, time.Minute)
		defer func() {
			err := client.Close()

			assert.Nil(t, err)
		}()

		flushRedis(t, client)

		storage := redis.NewStorage(client)

		actData, actTTL, err := storage.Get(ctx, url)

		assert.True(t, errors.Is(err, streaming.ErrDataNotFoundInStorage))
		assert.Equal(t, "", actData)
		assert.Equal(t, actTTL, time.Duration(0))
	})
	t.Run("get expired", func(t *testing.T) {
		client := redis.NewClient(host, db, time.Minute, time.Minute, time.Minute, 16, 16, time.Minute)
		defer func() {
			err := client.Close()

			assert.Nil(t, err)
		}()

		flushRedis(t, client)

		storage := redis.NewStorage(client)

		err := storage.Set(ctx, url, data, ttl)

		assert.Nil(t, err)

		// Wait until entry expires in Redis.
		time.Sleep(ttl + 100*time.Millisecond)

		actData, actTTL, err := storage.Get(ctx, url)

		assert.True(t, errors.Is(err, streaming.ErrDataNotFoundInStorage))
		assert.Equal(t, "", actData)
		assert.Equal(t, actTTL, time.Duration(0))
	})
}
