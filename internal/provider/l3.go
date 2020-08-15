package provider

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sony/gobreaker"

	"github.com/LasTshaMAN/streaming"
)

type L3 struct {
	minTTL time.Duration
	maxTTL time.Duration

	client *resty.Client

	cb *gobreaker.CircuitBreaker
	cbTimeout time.Duration
}

func NewL3(
	minTTL time.Duration,
	maxTTL time.Duration,
	client *resty.Client,
	cbTimeout time.Duration,
) *L3 {
	return &L3{
		minTTL: minTTL,
		maxTTL: maxTTL,
		client: client,
		cb: gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:          "L3 circuit breaker",
			Timeout:       cbTimeout,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > 0
			},
		}),
		cbTimeout: cbTimeout,
	}
}

// TODO
// Implement circuit-breaker to reduce the amount of requests to external services.

func (srv *L3) Get(_ context.Context, url string) (string, time.Duration, error) {
	// TODO
	// make sure resty closes response body

	respBody, err := srv.cb.Execute(func() (interface{}, error) {
		resp, err := srv.client.R().Get(url)
		if err != nil {
			return nil, fmt.Errorf("get data from URL, err: %w", err)
		}

		return resp.String(), nil
	})
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, gobreaker.ErrOpenState) {
		return "", srv.calculateTTL(), streaming.ErrDataCurrentlyUnavailable
	}
	if err != nil {
		return "", 0, fmt.Errorf("execute in circuit breaker, err: %w", err)
	}

	return respBody.(string), srv.calculateTTL(), nil
}

func (srv *L3) calculateTTL() time.Duration {
	return time.Duration(int64(srv.minTTL) + rand.Int63n(int64(srv.minTTL)+1))
}
