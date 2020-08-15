package provider

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/LasTshaMAN/streaming"
)

type L1 struct {
	storage streaming.DataStorage

	lock     sync.Locker
	fallback streaming.DataProvider

	// fallbackRoundTripTime is an upper estimate on the time it takes to fetch data from fallback provider.
	//
	// Fallback provider returns a ttl (time to live) for the data it provides.
	// We are using this ttl value (somewhat adjusted, as explained below) to find for what period of time we can safely cache the data
	// returned by Fallback provider.
	//
	// There is one caveat.
	// Fallback provider calculates this value at the moment of request serving, meaning that by the time we get its response
	// this ttl value is already stale.
	// fallbackRoundTripTime is an upper bound on the difference between ttl calculated by fallback provider and
	// ttl that is remaining when we get back fallback provider.
	fallbackRoundTripTime time.Duration
	// fallbackRoundTripTime is an upper estimate on the time it takes to write data to storage.
	// It is used in a similar way to how fallbackRoundTripTime is used.
	storageRoundTripTime time.Duration
}

func NewL1(
	storage streaming.DataStorage,
	fallback streaming.DataProvider,
	fallbackRoundTripTime time.Duration,
	storageRoundTripTime time.Duration,
) *L1 {
	return &L1{
		storage:               storage,
		fallback:              fallback,
		fallbackRoundTripTime: fallbackRoundTripTime,
		storageRoundTripTime:  storageRoundTripTime,
	}
}

func (srv *L1) Get(ctx context.Context, url string) (string, time.Duration, error) {
	data, ttl, err := srv.storage.Get(ctx, url)
	if err == nil {
		return data, ttl, nil
	}

	if !errors.Is(err, streaming.ErrDataNotFoundInStorage) {
		return "", 0, fmt.Errorf("get data from storage, err: %w", err)
	}

	// TODO
	// Split lock by URL
	srv.lock.Lock()
	defer srv.lock.Unlock()

	// Check once again whether data is in storage, since somebody must have put it there already while we
	// were performing "the fast scenario" (a piece of code above).

	data, ttl, err = srv.storage.Get(ctx, url)
	if err == nil {
		return data, ttl, nil
	}

	if !errors.Is(err, streaming.ErrDataNotFoundInStorage) {
		return "", 0, fmt.Errorf("get data from storage (under lock), err: %w", err)
	}

	// At this point its our job to fetch the data from fallback provider and cache it in our storage.

	data, ttl, err = srv.fallback.Get(ctx, url)
	if err != nil {
		return "", 0, fmt.Errorf("get data from fallback provider, err: %w", err)
	}

	ttl = srv.adjustTTL(ttl)

	err = srv.storage.Set(ctx, url, data, ttl)
	if err != nil {
		return "", 0, fmt.Errorf("set data (with expiration) for url in storage, err: %w", err)
	}

	return data, ttl, nil
}

// adjustTTL based on different factors.
func (srv *L1) adjustTTL(fallbackTTL time.Duration) time.Duration {
	// codeExecutionUpperEstimate estimates and adjustment for the code execution,
	// we need it because while we are executing this code fallbackTTL is getting even more out of date.
	//
	// This value is pretty much arbitrary, and might be adjusted in the future according to our needs.
	const codeExecutionUpperEstimate = 100 * time.Millisecond

	result := fallbackTTL - srv.fallbackRoundTripTime - srv.storageRoundTripTime - codeExecutionUpperEstimate

	if result < 0 {
		return 0
	}

	return result
}
