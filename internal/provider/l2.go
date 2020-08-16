package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/LasTshaMAN/streaming"
)

type L2 struct {
	logger log.Logger

	storage streaming.DataStorage

	locker   streaming.Locker
	fallback streaming.DataProvider
}

func NewL2(
	logger log.Logger,
	storage streaming.DataStorage,
	locker streaming.Locker,
	fallback streaming.DataProvider,
) *L2 {
	return &L2{
		logger:   logger,
		storage:  storage,
		fallback: fallback,
		locker:   locker,
	}
}

func (srv *L2) Get(ctx context.Context, url string) (data string, ttl time.Duration, err error) {
	data, ttl, err = srv.storage.Get(ctx, url)
	if err == nil {
		return data, ttl, nil
	}

	// TODO
	// Handle ErrDataCurrentlyUnavailable in L2

	if !errors.Is(err, streaming.ErrDataNotFoundInStorage) {
		return "", 0, fmt.Errorf("get data from storage, err: %w", err)
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

	_ = level.Info(srv.logger).Log("msg", "l2: fetch data from fallback provider")

	data, ttl, err = srv.fallback.Get(ctx, url)
	if err != nil {
		return "", 0, fmt.Errorf("get data from fallback provider, err: %w", err)
	}

	err = srv.storage.Set(ctx, url, data, ttl)
	if err != nil {
		return "", 0, fmt.Errorf("set data (with expiration) for url in storage, err: %w", err)
	}

	return data, ttl, nil
}
