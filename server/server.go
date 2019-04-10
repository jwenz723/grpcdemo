package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/jwenz723/grpcdemo/messaging"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"io"
	"net"
	"net/http"
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

// StreamResponse implements helloworld.GreeterServer
func (s *server) StreamMessages(stream messaging.MessagingService_StreamMessagesServer) error {
	n := &messaging.Message{Sender: "server", Message: "A streamed message from the server"}
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		if err := stream.Send(n); err != nil {
			return err
		}
	}
}

func (s *server) SendMessage(ctx context.Context, point *messaging.Message) (*messaging.Message, error) {
	n := &messaging.Message{Sender: "server", Message: "A single message from the server"}
	return n, nil
}

func newServer() *server {
	s := &server{}
	return s
}

var (
	port = flag.Int("port", 8080, "The server port")
)

func main() {
	flag.Parse()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Start prometheus metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2111", nil)
	}()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logger.Fatal("failed to listen",
			zap.Error(err))
	}

	logger.Info("starting grpc server",
		zap.Int("port", *port))

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_zap.UnaryServerInterceptor(logger),
			grpc_prometheus.UnaryServerInterceptor)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_zap.StreamServerInterceptor(logger),
			grpc_prometheus.StreamServerInterceptor)),
	)

	messaging.RegisterMessagingServiceServer(grpcServer, newServer())
	grpc_prometheus.Register(grpcServer)

	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("failed to start grpc server",
			zap.Error(err))
	}
}
