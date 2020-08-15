package redis_test

import (
	"sync"
	"testing"
	"time"

	"github.com/go-redsync/redsync"
	"github.com/stretchr/testify/assert"

	"github.com/LasTshaMAN/streaming/internal/redis"
)

func TestRedSync(t *testing.T) {
	t.Run("redsync must implement block-wait semantics, not try-error semantics", func(t *testing.T) {
		// We want to make sure redsync will block when mutex is occupied by other actors instead of returning an error.
		// This behavior is not described in their docs,
		// and it is faster to check it ourselves (in a test like this) than to dig through implementation.

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

		lock := redsync.New([]redsync.Pool{client}).NewMutex("some mutex", redsync.SetExpiry(time.Minute))

		f := func() {
			err := lock.Lock()
			assert.Nil(t, err)

			defer func() {
				success, unlockErr := lock.Unlock()
				assert.Nil(t, unlockErr)
				assert.True(t, success)
			}()

			time.Sleep(time.Millisecond)
		}

		wg := sync.WaitGroup{}

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				f()
				wg.Done()
			}()
		}

		wg.Wait()
	})
}
