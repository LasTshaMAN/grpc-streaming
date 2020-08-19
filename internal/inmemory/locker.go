package inmemory

import (
	"hash/maphash"
	"sync"
)

type Locker struct {
	locks []*sync.Mutex
	size  int

	// hSeed is a seed used for hashing algorithm to hash URLs.
	hSeed maphash.Seed
}

func NewLocker(size int) *Locker {
	locks := make([]*sync.Mutex, size)

	for i := 0; i < size; i++ {
		locks[i] = &sync.Mutex{}
	}

	return &Locker{
		locks: locks,
		size:  size,
		hSeed: maphash.MakeSeed(),
	}
}

func (locker *Locker) Lock(url string) error {
	lock := locker.getLock(url)

	lock.Lock()

	return nil
}

func (locker *Locker) Unlock(url string) (bool, error) {
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
