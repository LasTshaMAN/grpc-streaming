package random

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/LasTshaMAN/streaming"
)

type Service struct {
	provider streaming.DataProvider
	urls     []string
}

func NewService(urls []string, provider streaming.DataProvider) *Service {
	return &Service{
		urls:     urls,
		provider: provider,
	}
}

func (srv *Service) GetNext(ctx context.Context) (string, error) {
	idx := rand.Intn(len(srv.urls))

	url := srv.urls[idx]

	data, _, err := srv.provider.Get(ctx, url)
	if err != nil {
		return "", fmt.Errorf("get url, err: %w", err)
	}

	return data, nil
}
