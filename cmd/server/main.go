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
	"github.com/LasTshaMAN/streaming/internal/internet"
	"github.com/LasTshaMAN/streaming/internal/proxy"
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

	// TODO
	// rename all l1 and l2

	const (
		inetRequestTimeout        = 5 * time.Second
		inetDataUnavailablePeriod = 60 * time.Second

		//redisConnCount       = 16
		redisConnCount       = 1
		redisDialTimeout     = time.Second
		redisRequestTimeout  = time.Second
		redisIdleConnTimeout = 10 * time.Minute

		// TODO
		l2CodeExecutionUpperEstimate = 109 * time.Second

		// TODO
		lockerSize = 1
		//lockerSize = 100
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
	dLockExpiry := l2CodeExecutionUpperEstimate +
		redisDialTimeout + redisRequestTimeout +
		inetRequestTimeout +
		redisDialTimeout + redisRequestTimeout

	locker := redis.NewLocker(lockerSize, dLockExpiry, redisClient)

	inetClient := resty.NewWithClient(&http.Client{Timeout: inetRequestTimeout})

	// TODO
	//inetSimpleProvider := internet.NewSimpleProvider(logger, cfg.MinTimeout, cfg.MaxTimeout, inetDataUnavailablePeriod, inetClient)
	inetProvider := internet.NewProvider(logger, cfg.MinTimeout, cfg.MaxTimeout, inetDataUnavailablePeriod, inetClient)

	//redisInetProxy := proxy.NewProxy(
	//	logger,
	//	redisStorage,
	//	locker,
	//	inetSimpleProvider,
	//	// TODO
	//	func(fallbackTTL time.Duration) time.Duration {
	//		return fallbackTTL
	//	},
	//	//func(fallbackTTL time.Duration) time.Duration {
	//	//	// proxyCodeExecutionUpperEstimate estimates an adjustment for the code execution,
	//	//	// we need it because while we are executing code in this Proxy fallbackTTL is getting even more out of date.
	//	//	//
	//	//	// This value is pretty much arbitrary, and might be adjusted in the future according to our needs.
	//	//	const proxyCodeExecutionUpperEstimate = 100 * time.Millisecond
	//	//
	//	//	result := fallbackTTL - inetRequestTimeout - (redisDialTimeout + redisRequestTimeout) - proxyCodeExecutionUpperEstimate
	//	//
	//	//	if result < 0 {
	//	//		return 0
	//	//	}
	//	//
	//	//	return result
	//	//},
	//)
	redisInetProxy := proxy.NewProxy(
		logger,
		redisStorage,
		locker,
		inetProvider,
		// TODO
		func(fallbackTTL time.Duration) time.Duration {
			return fallbackTTL
		},
		//func(fallbackTTL time.Duration) time.Duration {
		//	// proxyCodeExecutionUpperEstimate estimates an adjustment for the code execution,
		//	// we need it because while we are executing code in this Proxy fallbackTTL is getting even more out of date.
		//	//
		//	// This value is pretty much arbitrary, and might be adjusted in the future according to our needs.
		//	const proxyCodeExecutionUpperEstimate = 100 * time.Millisecond
		//
		//	result := fallbackTTL - inetRequestTimeout - (redisDialTimeout + redisRequestTimeout) - proxyCodeExecutionUpperEstimate
		//
		//	if result < 0 {
		//		return 0
		//	}
		//
		//	return result
		//},
	)

	//inmemStorage :=
	//
	//inmemRedisProxy := provider.NewProxy(, redisInetProxy, inetRequestTimeout, redisDialTimeout + redisRequestTimeout)

	// TODO
	//randProvider := random.NewService(cfg.URLs, inetSimpleProvider)
	//randProvider := random.NewService(cfg.URLs, inmemRedisProxy)
	randProvider := random.NewLoggingMiddleware(logger, random.NewService(cfg.URLs, redisInetProxy))

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
