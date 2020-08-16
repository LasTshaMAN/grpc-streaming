package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"google.golang.org/grpc"

	gengrpc "github.com/LasTshaMAN/streaming/gen/grpc"
)

func main() {
	ctx := context.Background()

	var logger log.Logger
	{
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	const desiredConnectionsCnt = 1000

	// Gather some statistics to verify the solution validity, throughput, latency, ...
	var (
		replySuccessCnt = int64(0)
		replyFailureCnt = int64(0)

		connectedCnt = int64(0)

		startTime = time.Now()

		wg sync.WaitGroup
	)

	for i := 0; i < desiredConnectionsCnt; i++ {
		wg.Add(1)

		go func() {
			replySuccess, err := connect(ctx, "localhost:50051", logger)
			if err != nil {
				_ = level.Error(logger).Log("err", fmt.Errorf("connect, err: %w", err))
				return
			}

			if replySuccess {
				atomic.AddInt64(&replySuccessCnt, 1)
			} else {
				atomic.AddInt64(&replyFailureCnt, 1)
			}

			atomic.AddInt64(&connectedCnt, 1)
			if connectedCnt%(desiredConnectionsCnt/10) == 0 {
				deltaDuration := time.Now().Sub(startTime)

				msg := fmt.Sprintf("established connections: %d, took: %d ms", connectedCnt, deltaDuration.Milliseconds())
				_ = level.Info(logger).Log("mgs", msg)
			}

			wg.Done()
		}()
	}

	// Wait for all the reply success/failure data be gathered.
	wg.Wait()

	_ = level.Info(logger).Log("mgs", fmt.Sprintf("reply results, success: %d, failure: %d", replySuccessCnt, replyFailureCnt))

	// Just hang the process (not bothering with proper termination for now).
	select {}
}

// connect establishes a connection, makes sure it works properly and
// before returning this func spawns a background worker to handle the data stream coming on this connection.
// This background worker terminates upon a first error encountered.
func connect(ctx context.Context, target string, logger log.Logger) (firstReplySuccess bool, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)

	conn, err := grpc.DialContext(ctx, target, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return false, fmt.Errorf("dial target, err: %w", err)
	}

	c := gengrpc.NewStreamingServiceClient(conn)

	stream, err := c.GetRandomDataStream(context.Background(), &gengrpc.Request{})
	if err != nil {
		return false, fmt.Errorf("get random data stream, err: %w", err)
	}

	resp, err := stream.Recv()
	if err != nil {
		return false, fmt.Errorf("receive data from stream, err: %w", err)
	}

	err = ValidateReply(resp.Reply)
	if err != nil {
		return false, fmt.Errorf("validate reply, err: %w", err)
	}

	// At this point - consider the connection to be successfully established.

	go func() {
		process(stream, logger)

		closeErr := conn.Close()
		if closeErr != nil {
			_ = level.Error(logger).Log("err", fmt.Errorf("receive data from stream, err: %w", closeErr))
			return
		}

		cancel()
	}()

	const errReply = "err"

	return resp.Reply != errReply, nil
}

func process(stream gengrpc.StreamingService_GetRandomDataStreamClient, logger log.Logger) {
	for {
		resp, err := stream.Recv()
		if err != nil {
			_ = level.Error(logger).Log("err", fmt.Errorf("receive data from stream, err: %w", err))
			return
		}

		err = ValidateReply(resp.Reply)
		if err != nil {
			_ = level.Error(logger).Log("err", fmt.Errorf("validate reply, err: %w", err))
			return
		}
	}
}
