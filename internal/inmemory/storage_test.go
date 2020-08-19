package inmemory_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/LasTshaMAN/streaming"
	"github.com/LasTshaMAN/streaming/internal/inmemory"
)

func TestStorage(t *testing.T) {
	const (
		url           = "some url"
		data          = "some data"
		ttl           = 2 * time.Second
		deltaDuration = time.Millisecond
	)

	var (
		ctx = context.Background()

		now = time.Now()
	)

	t.Run("get existent", func(t *testing.T) {
		storage := inmemory.NewStorage(func() func() time.Time {
			var (
				calls = []time.Time{
					now,
					now.Add(deltaDuration),
				}
				idx = 0
			)

			return func() time.Time {
				defer func() {
					idx++
				}()

				return calls[idx]
			}
		}())

		err := storage.Set(ctx, url, data, ttl)

		assert.Nil(t, err)

		actData, actTTL, err := storage.Get(ctx, url)

		assert.Nil(t, err)
		assert.Equal(t, data, actData)
		assert.Equal(t, now.Add(ttl).Sub(now.Add(deltaDuration)), actTTL)
	})
	t.Run("get non-existent", func(t *testing.T) {
		storage := inmemory.NewStorage(nil)

		actData, actTTL, err := storage.Get(ctx, url)

		assert.True(t, errors.Is(err, streaming.ErrDataNotFoundInStorage))
		assert.Equal(t, "", actData)
		assert.Equal(t, actTTL, time.Duration(0))
	})
	t.Run("get expired", func(t *testing.T) {
		storage := inmemory.NewStorage(func() func() time.Time {
			var (
				calls = []time.Time{
					now,
					now.Add(ttl + deltaDuration),
				}
				idx = 0
			)

			return func() time.Time {
				defer func() {
					idx++
				}()

				return calls[idx]
			}
		}())

		err := storage.Set(ctx, url, data, ttl)

		assert.Nil(t, err)

		actData, actTTL, err := storage.Get(ctx, url)

		assert.True(t, errors.Is(err, streaming.ErrDataNotFoundInStorage))
		assert.Equal(t, "", actData)
		assert.Equal(t, actTTL, time.Duration(0))
	})
}
