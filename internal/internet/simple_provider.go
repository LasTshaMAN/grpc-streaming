package internet

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-resty/resty/v2"

	"github.com/LasTshaMAN/streaming"
)

type SimpleProvider struct {
	logger log.Logger

	minTTL time.Duration
	maxTTL time.Duration

	dataUnavailablePeriod time.Duration

	client *resty.Client
}

func NewSimpleProvider(
	logger log.Logger,
	minTTL time.Duration,
	maxTTL time.Duration,
	dataUnavailablePeriod time.Duration,
	client *resty.Client,
) *SimpleProvider {
	return &SimpleProvider{
		logger:                logger,
		minTTL:                minTTL,
		maxTTL:                maxTTL,
		dataUnavailablePeriod: dataUnavailablePeriod,
		client:                client,
	}
}

func (srv *SimpleProvider) Get(_ context.Context, url string) (string, time.Duration, error) {
	// TODO
	// make sure resty closes response body

	resp, err := srv.client.R().Get(url)
	if err != nil {
		_ = level.Error(srv.logger).Log("err", fmt.Errorf("get data from URL: %s, err: %w", url, err))

		return "", srv.dataUnavailablePeriod, streaming.ErrDataCurrentlyUnavailable
	}

	return resp.String(), srv.calculateTTL(), nil
}

func (srv *SimpleProvider) calculateTTL() time.Duration {
	return time.Duration(int64(srv.minTTL) + rand.Int63n(int64(srv.minTTL)+1))
}
