package grpc_server

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
)

type Option func(*Options)

type Options struct {
	Addr      string
	TLSConfig *struct {
		CertFile string
		KeyFile  string
	}
	UnaryInts    []grpc.UnaryServerInterceptor
	StreamInts   []grpc.StreamServerInterceptor
	StatsHandler stats.Handler
	MaxRecvBytes int
	MaxSendBytes int
}

func WithAddr(addr string) Option {
	return func(o *Options) {
		o.Addr = addr
	}
}

func WithTLS(certFile, keyFile string) Option {
	return func(o *Options) {
		o.TLSConfig = &struct {
			CertFile string
			KeyFile  string
		}{
			CertFile: certFile,
			KeyFile:  keyFile,
		}
	}
}

func WithUnaryUse(i grpc.UnaryServerInterceptor) Option {
	return func(o *Options) {
		o.UnaryInts = append(o.UnaryInts, i)
	}
}

func WithStreamUse(i grpc.StreamServerInterceptor) Option {
	return func(o *Options) {
		o.StreamInts = append(o.StreamInts, i)
	}
}

func WithMaxRecvBytes(maxRecvBytes int) Option {
	return func(o *Options) {
		o.MaxRecvBytes = maxRecvBytes
	}
}

func WithMaxSendBytes(maxSendBytes int) Option {
	return func(o *Options) {
		o.MaxSendBytes = maxSendBytes
	}
}

func WithStatsHandler(statsHandler stats.Handler) Option {
	return func(o *Options) {
		o.StatsHandler = statsHandler
	}
}