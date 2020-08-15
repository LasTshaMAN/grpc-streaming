package streaming

import (
	"context"
	"time"
)

type RandomDataProvider interface {
	GetNext(context.Context) (data string, err error)
}

type DataProvider interface {
	Get(ctx context.Context, url string) (data string, ttl time.Duration, err error)
}

// DataStorage provides key-value storage functionality for data.
//
// DataStorage can be safely used concurrently from multiple go-routines.
type DataStorage interface {
	Get(ctx context.Context, url string) (data string, ttl time.Duration, err error)
	Set(ctx context.Context, url string, data string, ttl time.Duration) error
}

// DistributedLock is a lock that resides somewhere outside of the process this service is running in,
// it is used to synchronize access to some resource between several processes
// (not just go-routines withing this process) running concurrently.
type DistributedLock interface {
	// Lock acquires lock.
	Lock() error
	// Unlock previously acquired lock.
	Unlock() (bool, error)
}
