package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/jwenz723/grpcdemo/messaging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"io"
	"net"
	"net/http"
        "math/big"
)

type server struct {
	uLogger *zap.Logger
	sLogger *zap.Logger
}

func (s server) BadFunc() {
        s.sLogger = nil
}

func (s *server) StreamMessages(stream messaging.MessagingService_StreamMessagesServer) error {
	n := &messaging.Message{Sender: "server", Message: "A streamed message from the server"}
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		err = stream.Send(n)
		grpcMessagesSent.WithLabelValues("stream").Inc()
		s.sLogger.Info("sent")
	}
}

func (s *server) SendMessage(ctx context.Context, point *messaging.Message) (*messaging.Message, error) {
	n := &messaging.Message{Sender: "server", Message: "A single message from the server"}
	grpcMessagesSent.WithLabelValues("unary").Inc()
	s.uLogger.Info("sent")
	return n, nil
}

func newServer(logger *zap.Logger) *server {
	s := &server{
		sLogger: logger.With(zap.String("type", "stream")),
		uLogger: logger.With(zap.String("type", "unary")),
	}
	return s
}

var (
	port             = flag.Int("port", 8080, "The server port")
	grpcMessagesSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "grpc_messages_sent",
			Help:        "The total number of grpc messages sent",
			ConstLabels: prometheus.Labels{"from": "server"},
		},
		[]string{"method"},
	)
)

func superfluous() bool {
        return true
}

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(grpcMessagesSent)
}

func main() {
	flag.Parse()

        if superfluous() {
                fmt.Println("something")
        }
        username := "testing"
        password := "something secret"
        fmt.Println(username + password)

        if username == "Tom" {
	        superfluous()
        } else if username == "Tom" {
	        superfluous()
        }

        r := new(big.Rat)
	r.SetString("355/113")
	fmt.Println(r.FloatString(3))

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

	grpcServer := grpc.NewServer()
	messaging.RegisterMessagingServiceServer(grpcServer, newServer(logger))

	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("failed to start grpc server",
			zap.Error(err))
	}
}
