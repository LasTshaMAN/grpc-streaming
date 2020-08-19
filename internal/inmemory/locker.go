package inmemory

import (
	"hash/maphash"
	"sync"

	"github.com/go-kit/kit/log"
)

type Locker struct {
	// TODO
	logger log.Logger

	locks []*sync.Mutex
	size  int

	// hSeed is a seed used for hashing algorithm to hash URLs.
	hSeed maphash.Seed
}

func NewLocker(logger log.Logger, size int) *Locker {
	locks := make([]*sync.Mutex, size)

	for i := 0; i < size; i++ {
		locks[i] = &sync.Mutex{}
	}

	return &Locker{
		logger: logger,
		locks: locks,
		size:  size,
		hSeed: maphash.MakeSeed(),
	}
}

func (locker *Locker) Lock(url string) error {
	// TODO
	//_ = level.Error(locker.logger).Log("err", fmt.Sprintf("(locker *Locker) Lock, url: %s", url))

	lock := locker.getLock(url)

	lock.Lock()

	return nil
}

func (locker *Locker) Unlock(url string) (bool, error) {
	// TODO
	//_ = level.Error(locker.logger).Log("err", fmt.Sprintf("(locker *Locker) Unlock, url: %s", url))

	lock := locker.getLock(url)

	lock.Unlock()

	return true, nil
}

func (locker *Locker) getLock(url string) *sync.Mutex {
	var h maphash.Hash
	h.SetSeed(locker.hSeed)

	_, _ = h.WriteString(url)
	defer h.Reset()

	hash := int(h.Sum64() >> 33)

	return locker.locks[hash%locker.size]
}
