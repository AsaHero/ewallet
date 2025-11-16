package grpc_server

import (
	"context"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	*grpc.Server
	addr string
	lis  net.Listener
}

func New(opts ...Option) (*Server, error) {
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

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(o.UnaryInts...),
		grpc.ChainStreamInterceptor(o.StreamInts...),
		grpc.MaxRecvMsgSize(o.MaxRecvBytes),
		grpc.MaxSendMsgSize(o.MaxSendBytes),
		grpc.StatsHandler(o.StatsHandler),
	)

	lis, err := net.Listen("tcp", o.Addr)
	if err != nil {
		return nil, err
	}

	return &Server{
		Server: s,
		addr:   o.Addr,
		lis:    lis,
	}, nil
}

func (s *Server) Run() error {
	return s.Serve(s.lis)
}

func (s *Server) Stop(ctx context.Context) error {
	stop := make(chan struct{})

	go func() {
		s.GracefulStop()
		close(stop)
	}()

	select {
	case <-stop:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
