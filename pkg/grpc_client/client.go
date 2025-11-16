package grpc_client

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	*grpc.ClientConn
	addr string
}

func New(opts ...Option) (*Client, error) {
	o := &Options{
		Addr:         "0.0.0.0:50051",
		TLSConfig:    nil,
		UnaryInts:    nil,
		StreamInts:   nil,
		MaxRecvBytes: 1 << 20,
		MaxSendBytes: 1 << 20,
	}

	for _, opt := range opts {
		opt(o)
	}

	c, err := grpc.NewClient(
		o.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(o.UnaryInts...),
		grpc.WithChainStreamInterceptor(o.StreamInts...),
		grpc.WithStatsHandler(o.StatsHandler),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallSendMsgSize(o.MaxSendBytes),
			grpc.MaxCallRecvMsgSize(o.MaxRecvBytes),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client connection: %w", err)
	}

	return &Client{
		ClientConn: c,
		addr:       o.Addr,
	}, nil
}

func (s *Client) Close() error {
	return s.ClientConn.Close()
}
