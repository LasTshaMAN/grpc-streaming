package proxy

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/LasTshaMAN/streaming"
)

type Proxy struct {
	logger log.Logger

	storage streaming.DataStorage

	locker streaming.Locker

	fallback streaming.DataProvider

	// adjustTTL based on different factors (these factors are defined by the user of this struct -> hence this is a func).
	//
	// Fallback provider returns a ttl (time to live) for the data it provides.
	// We are using this ttl value (somewhat adjusted, as explained below) to find for what period of time we can safely cache the data
	// returned by Fallback provider,
	// and then we can return this cached data to our caller instead of calling fallback provider (that is slower than we are).
	//
	// There is one caveat.
	// Fallback provider calculates ttl value at the moment of request serving, meaning that by the time we get its response
	// this ttl value is already stale.
	// That's why we need to adjust the ttl value (to avoid serving stale data) fallback provider returns.
	adjustTTL func(fallbackTTL time.Duration) time.Duration
}

func NewProxy(
	logger log.Logger,
	storage streaming.DataStorage,
	locker streaming.Locker,
	fallback streaming.DataProvider,
	adjustTTL func(fallbackTTL time.Duration) time.Duration,
) *Proxy {
	return &Proxy{
		logger:    logger,
		storage:   storage,
		fallback:  fallback,
		locker:    locker,
		adjustTTL: adjustTTL,
	}
}

func (srv *Proxy) Get(ctx context.Context, url string) (string, time.Duration, error) {
	found, data, ttl, err := srv.tryStorage(ctx, url)
	if found {
		return data, ttl, err
	}
	if err != nil {
		return "", 0, fmt.Errorf("try storage, err: %w", err)
	}

	err = srv.locker.Lock(url)
	if err != nil {
		return "", 0, fmt.Errorf("lock locker, err: %w", err)
	}
	defer func() {
		success, unlockErr := srv.locker.Unlock(url)
		if unlockErr != nil {
			_ = level.Error(srv.logger).Log("err", fmt.Errorf("unlock locker, err: %w", unlockErr))
			return
		}
		if !success {
			_ = level.Error(srv.logger).Log("err", "release lock, operation is unsuccessful")
			return
		}
	}()

	// Check once again whether the data is in storage, since another go-routine might have put it there while we
	// were performing "the fast scenario" (a piece of code above).

	found, data, ttl, err = srv.tryStorage(ctx, url)
	if found {
		return data, ttl, err
	}
	if err != nil {
		return "", 0, fmt.Errorf("try storage, err: %w", err)
	}

	// At this point nobody concurrently with us can to fetch the data from fallback provider and cache it in our storage.

	_ = level.Info(srv.logger).Log("msg", fmt.Sprintf("Proxy: fetch data from fallback provider, URL: %s", url))

	data, ttl, err = srv.fallback.Get(ctx, url)
	if err != nil && !errors.Is(err, streaming.ErrDataCurrentlyUnavailable) {
		return "", 0, fmt.Errorf("get data from fallback provider, err: %w", err)
	}

	if errors.Is(err, streaming.ErrDataCurrentlyUnavailable) {
		// We are caching "temporary unavailable" error response for efficiency / performance reasons.

		ttl = srv.adjustTTL(ttl)

		setErr := srv.storage.Set(ctx, url, dataUnavailableMarker, ttl)
		if setErr != nil {
			return "", 0, fmt.Errorf("set data unavailable marker (with expiration) for url in storage, err: %w", setErr)
		}

		return "", ttl, streaming.ErrDataCurrentlyUnavailable
	}

	ttl = srv.adjustTTL(ttl)

	setErr := srv.storage.Set(ctx, url, data, ttl)
	if setErr != nil {
		return "", 0, fmt.Errorf("set data (with expiration) for url in storage, err: %w", setErr)
	}

	return data, ttl, nil
}

func (srv *Proxy) tryStorage(ctx context.Context, url string) (found bool, data string, ttl time.Duration, err error) {
	data, ttl, err = srv.storage.Get(ctx, url)

	if err != nil && !errors.Is(err, streaming.ErrDataNotFoundInStorage) {
		return false, "", 0, fmt.Errorf("get data from storage, err: %w", err)
	}

	if errors.Is(err, streaming.ErrDataNotFoundInStorage) {
		return false, "", 0, nil
	}

	if data == dataUnavailableMarker {
		return true, "", ttl, streaming.ErrDataCurrentlyUnavailable
	}

	return true, data, ttl, nil
}

const (
	dataUnavailableMarker = "data is currently unavailable"
)
