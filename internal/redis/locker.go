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

	// TODO
	// In order to get the same hash for the same URL in two different processes (running this service)
	// we need to use the same maphash.Seed.
	// From maphash.Seed documentation:
	//   ...
	//   Each Seed value is local to a single process and cannot be serialized
	//   or otherwise recreated in a different process.
	// Thus we can't use maphash.Seed & maphash.Hash when running multiple instances of this service and need to move to find else.
	//
	// hSeed is a seed used for hashing algorithm to hash URLs.
	hSeed maphash.Seed
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
		hSeed: maphash.MakeSeed(),
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
	var h maphash.Hash
	h.SetSeed(locker.hSeed)

	_, _ = h.WriteString(url)
	defer h.Reset()

	hash := int(h.Sum64() >> 33)

	return locker.locks[hash%locker.size]
}
