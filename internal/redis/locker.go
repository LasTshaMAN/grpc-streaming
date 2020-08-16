package redis

import (
	"fmt"
	"hash/maphash"
	"time"

	"github.com/go-redsync/redsync"
	"github.com/gomodule/redigo/redis"
)

type Locker struct {
	pool *redis.Pool

	locks []*redsync.Mutex
	size  int

	hasher maphash.Hash
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
	_, _ = locker.hasher.WriteString(url)
	defer locker.hasher.Reset()

	hash := int(locker.hasher.Sum64() >> 33)

	return locker.locks[hash%locker.size]
}
