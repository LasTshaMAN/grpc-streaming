package inmemory

import (
	"hash/fnv"
	"sync"
)

type Locker struct {
	locks []*sync.Mutex
	size  int
}

func NewLocker(size int) *Locker {
	locks := make([]*sync.Mutex, size)

	for i := 0; i < size; i++ {
		locks[i] = &sync.Mutex{}
	}

	return &Locker{
		locks: locks,
		size:  size,
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
	h := fnv.New64()

	_, _ = h.Write([]byte(url))
	defer h.Reset()

	hash := h.Sum64()

	return locker.locks[hash%uint64(locker.size)]
}
