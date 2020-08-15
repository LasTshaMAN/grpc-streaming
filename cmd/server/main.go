package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-resty/resty/v2"
	"google.golang.org/grpc"

	gengrpc "github.com/LasTshaMAN/streaming/gen/grpc"
	"github.com/LasTshaMAN/streaming/internal/api"
	"github.com/LasTshaMAN/streaming/internal/config"
	"github.com/LasTshaMAN/streaming/internal/provider"
	"github.com/LasTshaMAN/streaming/internal/random"
	"github.com/LasTshaMAN/streaming/internal/redis"
)

// TODO
// check that all parameters in funcs makes sense

func main() {
	var logger log.Logger
	{
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	cfg, err := config.Parse("config/config.yml")
	if err != nil {
		_ = level.Error(logger).Log("err", fmt.Errorf("parse config file, err: %w", err))
		return
	}

	const (
		l3RequestTimeout      = 5 * time.Second
		l3CircuitBreakerTimeout = 60 * time.Second


		redisConnCount = 16
		redisDialTimeout = time.Second
		redisRequestTimeout = time.Second
		redisIdleConnTimeout = 10 * time.Minute

		l2CodeExecutionUpperEstimate = time.Second

		lockerSize = 100
	)

	redisClient := redis.NewClient(
		"localhost:6379",
		0,
		redisDialTimeout,
		redisRequestTimeout,
		redisRequestTimeout,
		redisConnCount,
		redisConnCount,
		redisIdleConnTimeout,
	)
	defer func() {
		err := redisClient.Close()
		if err != nil {
			_ = level.Error(logger).Log("err", fmt.Errorf("close Redis client, err: %w", err))
		}
	}()

	redisStorage := redis.NewStorage(redisClient)

	// We need to make sure distributed lock won't expire before the protected section of code finishes its execution.
	// Also, we don't want distributed lock to be held for longer than necessary (cause that might affect service availability).
	// Thus, we are defining dLockExpiry below based on these considerations.
	dLockExpiry :=  l2CodeExecutionUpperEstimate +
		redisDialTimeout + redisRequestTimeout +
		l3RequestTimeout +
		redisDialTimeout + redisRequestTimeout

	locker := redis.NewLocker(lockerSize, dLockExpiry, redisClient)

	l3Client := resty.NewWithClient(&http.Client{Timeout: l3RequestTimeout})

	l3Provider := provider.NewL3(cfg.MinTimeout, cfg.MaxTimeout, l3Client, l3CircuitBreakerTimeout)

	l2Provider := provider.NewL2(logger, redisStorage, locker, l3Provider)

	//inmemStorage :=
	//
	//l1Provider := provider.NewL1(, l2Provider, l3RequestTimeout, redisDialTimeout + redisRequestTimeout)

	// TODO
	//randProvider := random.NewService(cfg.URLs, l3Provider)
	randProvider := random.NewService(cfg.URLs, l2Provider)
	//randProvider := random.NewService(cfg.URLs, l1Provider)

	server := api.NewServer(logger, cfg.NumberOfRequests, randProvider)

	grpcServer := grpc.NewServer()
	defer grpcServer.GracefulStop()

	gengrpc.RegisterStreamingServiceServer(grpcServer, server)

	conn, err := net.Listen("tcp", ":50051")
	if err != nil {
		_ = level.Error(logger).Log("err", fmt.Errorf("listen, err: %w", err))
		return
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			_ = level.Error(logger).Log("err", fmt.Errorf("close connection, err: %w", err))
		}
	}()

	if err := grpcServer.Serve(conn); err != nil {
		_ = level.Error(logger).Log("err", fmt.Errorf("serve, err: %w", err))
		return
	}
}
