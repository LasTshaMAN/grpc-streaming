package inmemory

import (
	"context"
	"sync"
	"time"

	"github.com/LasTshaMAN/streaming"
)

type Storage struct {
	storage sync.Map

	now func() time.Time
}

func NewStorage(now func() time.Time) *Storage {
	return &Storage{
		now: now,
	}
}

func (s *Storage) Get(_ context.Context, url string) (string, time.Duration, error) {
	eObj, ok := s.storage.Load(url)
	if !ok {
		return "", 0, streaming.ErrDataNotFoundInStorage
	}

	e := eObj.(entry)

	expiresAt := e.createdAt.Add(e.ttl)
	ttl := expiresAt.Sub(s.now())

	if ttl < 0 {
		return "", 0, streaming.ErrDataNotFoundInStorage
	}

	return e.data, ttl, nil
}

func (s *Storage) Set(_ context.Context, url string, data string, ttl time.Duration) error {
	e := entry{
		data:      data,
		createdAt: s.now(),
		ttl:       ttl,
	}

	s.storage.Store(url, e)

	return nil
}

type entry struct {
	data      string
	createdAt time.Time
	ttl       time.Duration
}
