// Package streaming describes the domain model for this service.
package streaming

import (
	"context"
	"time"
)

// RandomDataProvider is designed to provide random data.
type RandomDataProvider interface {
	// GetNext returns the next chunk of random data.
	GetNext(context.Context) (data string, err error)
}

// DataProvider is designed to provide data identified by URL.
//
// Data providers can be layered on top of each other so that faster ones can serve as caches for slower ones.
//
// DataProvider can be safely used concurrently from multiple go-routines.
type DataProvider interface {
	// Get returns data identified by url, with a certain ttl (time to live) duration after which this data is considered to be stale.
	//
	// When ErrDataCurrentlyUnavailable is returned, ttl might have non-zero value,
	// in this case provider will continue to return ErrDataCurrentlyUnavailable for ttl duration
	// basically allowing for caching ErrDataCurrentlyUnavailable.
	Get(ctx context.Context, url string) (data string, ttl time.Duration, err error)
}

// TempDataStorage provides key-value storage to store the data this service works with.
// It is a "temporary" storage, meaning that whatever we store in it has an expiration time and 
// will disappear from storage after this time elapses.
//
// TempDataStorage can be safely used concurrently from multiple go-routines.
type TempDataStorage interface {
	// Get returns data identified by url, with a certain ttl (time to live) duration after which this data
	// might disappear from this storage.
	Get(ctx context.Context, url string) (data string, ttl time.Duration, err error)
	// Set stores data identified by url within this data storage for ttl period.
	Set(ctx context.Context, url string, data string, ttl time.Duration) error
}

// Locker manages locks each of which is identified by a URL.
//
// For every two URLs && url1 == url2 Locker methods must operate on the same lock,
// but when url1 != url2, Locker might or might not operate on the same lock (it is up to the implementation).
//
// Locker can be safely used concurrently from multiple go-routines.
type Locker interface {
	// Lock acquires a lock (associated with the provided url).
	Lock(url string) error
	// Unlock previously acquired lock (associated with the provided url).
	Unlock(url string) (bool, error)
}
