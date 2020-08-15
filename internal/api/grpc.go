package api

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/LasTshaMAN/streaming"
	gengrpc "github.com/LasTshaMAN/streaming/gen/grpc"
)

type Server struct {
	gengrpc.UnimplementedStreamingServiceServer

	logger log.Logger

	randReqNumber int

	provider streaming.RandomDataProvider
}

func NewServer(logger log.Logger, randReqNumber int, provider streaming.RandomDataProvider) *Server {
	return &Server{
		logger: logger,
		randReqNumber: randReqNumber,
		provider:      provider,
	}
}

func (srv *Server) GetRandomDataStream(
	_ *gengrpc.Request,
	stream gengrpc.StreamingService_GetRandomDataStreamServer,
) error {
	_ = level.Info(srv.logger).Log("msg", "received a request")

	for i := 0; i < 3; i++ {
		data, err := srv.provider.GetNext(stream.Context())
		if err != nil {
			_ = level.Error(srv.logger).Log("err", fmt.Errorf("get next random data, err: %w", err))

			data = "err"
		}
		resp := &gengrpc.Response{
			Reply: data,
		}
		err = stream.Send(resp)
		if err != nil {
			return fmt.Errorf("send response, err: %w", err)
		}
	}

	// Hold on to connection just to prove that we can.
	select {}
}
