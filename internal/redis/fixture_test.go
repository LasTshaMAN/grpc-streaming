package redis_test

import (
	"context"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
)

// flushRedis erases all the data we have in Redis instance.
//
// This solution if far from ideal,
// because it will erase all the data (possibly something valuable we are storing on our local machine)
// in the Redis instance we've connected to,
// and as a consequence - we won't be able to run tests relying on this func in parallel.
// But it will work fine for now.
func flushRedis(t *testing.T, client *redis.Pool) {
	conn, err := client.GetContext(context.Background())
	defer func() {
		closeErr := conn.Close()

		assert.Nil(t, closeErr)
	}()

	assert.Nil(t, err)

	_, err = conn.Do("FLUSHALL")

	assert.Nil(t, err)
}
