package grpc

import (
	"context"
	// "log"
	"net"
	"os"
	"os/signal"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/kenanya/jt-video-bank/pkg/api/v1"
	"github.com/kenanya/jt-video-bank/pkg/logger"
	"github.com/kenanya/jt-video-bank/pkg/protocol/grpc/middleware"
)

func RunServer(ctx context.Context, v1API v1.VideoBankServiceServer, port string) error {
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	// gRPC server statup options
	opts := []grpc.ServerOption{}

	// add middleware
	opts = middleware.AddLogging(logger.Log, opts)

	// server := grpc.NewServer()
	server := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,           // <--- This fixes it!
		}),
	)
	v1.RegisterVideoBankServiceServer(server, v1API)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			// log.Println("shutting down gRPC server...")
			logger.Log.Warn("shutting down gRPC server...")
			server.GracefulStop()
			<-ctx.Done()
		}
	}()

	// log.Println("starting gRPC server...")
	logger.Log.Info("starting gRPC server...")
	return server.Serve(listen)
}