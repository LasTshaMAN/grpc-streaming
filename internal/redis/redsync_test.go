package redis_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/go-redsync/redsync"
	"github.com/stretchr/testify/assert"

	"github.com/LasTshaMAN/streaming/internal/redis"
)

func TestRedSync(t *testing.T) {
	t.Run("redsync must not block forever", func(t *testing.T) {
		// We want to make sure redsync Lock will not block an actor on mutex forever when mutex is not available.
		// This behavior is not described in their docs.

		client := redis.NewClient(
			"localhost:6379",
			0,
			time.Minute,
			time.Minute,
			time.Minute,
			16,
			16,
			time.Minute,
		)

		lock := redsync.New([]redsync.Pool{client}).NewMutex(
			"some mutex",
			redsync.SetExpiry(time.Minute),
			redsync.SetTries(1),
			redsync.SetRetryDelay(time.Millisecond),
		)

		err := lock.Lock()
		assert.Nil(t, err, err)

		wg := sync.WaitGroup{}

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				err := lock.Lock()

				assert.True(t, errors.Is(err, redsync.ErrFailed), err)

				wg.Done()
			}()
		}

		// Occupy lock for a long time.
		time.Sleep(100 * time.Millisecond)

		success, unlockErr := lock.Unlock()
		assert.Nil(t, unlockErr)
		assert.True(t, success)

		wg.Wait()
	})
}
