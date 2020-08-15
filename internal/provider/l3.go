package provider

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-resty/resty/v2"
)

type L3 struct {
	minTTL time.Duration
	maxTTL time.Duration

	client *resty.Client
}

func NewL3(
	minTTL time.Duration,
	maxTTL time.Duration,
	client *resty.Client,
) *L3 {
	return &L3{
		client: client,
		minTTL: minTTL,
		maxTTL: maxTTL,
	}
}

// TODO
// Implement circuit-breaker to reduce the amount of requests to external services.

func (srv *L3) Get(_ context.Context, url string) (string, time.Duration, error) {
	resp, err := srv.client.R().Get(url)
	if err != nil {
		return "", 0, fmt.Errorf("get data from URL, err: %w", err)
	}

	ttl := int64(srv.minTTL) + rand.Int63n(int64(srv.minTTL)+1)

	return resp.String(), time.Duration(ttl), nil
}
