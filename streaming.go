// Package streaming describes the domain model for this service.
package streaming

import (
	"context"
	"time"
)

type RandomDataProvider interface {
	GetNext(context.Context) (data string, err error)
}

type DataProvider interface {
	// Get returns data by url, with a certain ttl (time to live) duration after which this data is considered to be stale.
	Get(ctx context.Context, url string) (data string, ttl time.Duration, err error)
}

// DataStorage provides key-value storage functionality for data.
//
// DataStorage can be safely used concurrently from multiple go-routines.
type DataStorage interface {
	Get(ctx context.Context, url string) (data string, ttl time.Duration, err error)
	Set(ctx context.Context, url string, data string, ttl time.Duration) error
}

type Locker interface {
	// Lock acquires a lock (associated with the provided url).
	Lock(url string) error
	// Unlock previously acquired lock (associated with the provided url).
	Unlock(url string) (bool, error)
}
