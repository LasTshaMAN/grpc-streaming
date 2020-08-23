package redis

import (
	"fmt"
	"hash/fnv"
	"time"

	"github.com/go-redsync/redsync"
	"github.com/gomodule/redigo/redis"
)

type Locker struct {
	pool *redis.Pool

	locks []*redsync.Mutex
	size  int
}

func NewLocker(size int, lockExpiry time.Duration, pool *redis.Pool) *Locker {
	r := redsync.New([]redsync.Pool{pool})

	locks := make([]*redsync.Mutex, size)
	for i := 0; i < size; i++ {
		locks[i] = r.NewMutex(fmt.Sprintf("lock %d", i), redsync.SetExpiry(lockExpiry))
	}

	return &Locker{
		pool:  pool,
		locks: locks,
		size:  size,
	}
}

func (locker *Locker) Lock(url string) error {
	lock := locker.getLock(url)

	return lock.Lock()
}

func (locker *Locker) Unlock(url string) (bool, error) {
	lock := locker.getLock(url)

	return lock.Unlock()
}

func (locker *Locker) getLock(url string) *redsync.Mutex {
	h := fnv.New64()

	_, _ = h.Write([]byte(url))
	defer h.Reset()

	hash := h.Sum64()

	return locker.locks[hash%uint64(locker.size)]
}
