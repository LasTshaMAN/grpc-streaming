package random

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type LoggingMiddleware struct {
	logger log.Logger
	srv    *Service
}

// NewLoggingMiddleware wraps service with logging middleware.
func NewLoggingMiddleware(logger log.Logger, srv *Service) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
		srv:    srv,
	}
}

func (mw *LoggingMiddleware) GetNext(ctx context.Context) (string, error) {
	defer func(begin time.Time) {
		_ = level.Info(mw.logger).Log(
			"method", "GetNext",
			"took", time.Since(begin),
		)
	}(time.Now())

	return mw.srv.GetNext(ctx)
}
