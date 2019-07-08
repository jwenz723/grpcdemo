package main

import (
	"context"
	"flag"
	"github.com/jwenz723/grpcdemo/messaging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"io"
	"net/http"
	"os"
	"time"
)

var (
	serverAddrFlag = flag.String("server_addr", "", "The server address in the format of host:port")
	useStreaming   = flag.Bool("use_streaming", false, "Setting this will use grpc streaming instead of repeated single messages")
	waitNanos      = flag.Int("wait_nanos", 500, "The number of nanoseconds to wait before sending messages (this applies to both single and stream messages)")
	waitDuration   time.Duration
	grpcMessagesSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "grpc_messages_sent",
			Help:        "The total number of grpc messages sent",
			ConstLabels: prometheus.Labels{"from": "client"},
		},
		[]string{"method"},
	)
)

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(grpcMessagesSent)
}

func main() {
	flag.Parse()

	grpcSendType := "unary"
	if *useStreaming {
		grpcSendType = "stream"
	}

	logger, _ := zap.NewProduction()
	logger = logger.With(zap.String("type", grpcSendType))
	defer logger.Sync()

	serverAddr := "localhost:8080"
	waitDuration = time.Duration(*waitNanos) * time.Nanosecond

	if *serverAddrFlag != "" {
		serverAddr = *serverAddrFlag
	} else if serverAddrEnv, ok := os.LookupEnv("SERVER_ADDR"); ok {
		serverAddr = serverAddrEnv
	}

	// Start prometheus metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

	// Note: load balancing between servers requires an L7 load balancer
	// (like linkerd) between this client and the servers
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())

	if err != nil {
		logger.Fatal("fail to dial",
			zap.Error(err))
	}
	defer conn.Close()

	client := messaging.NewMessagingServiceClient(conn)

	if *useStreaming {
		waitc := make(chan struct{})
		stream, err := client.StreamMessages(context.Background())
		if err != nil {
			logger.Fatal("error opening stream",
				zap.Error(err))
		}
		defer stream.CloseSend()

		n := &messaging.Message{Sender: "client", Message: "A streamed message from a client"}
		go func() {
			for {
				_, err := stream.Recv()
				if err == io.EOF {
					// read done.
					close(waitc)
					return
				}
				if err != nil {
					logger.Fatal("error receiving message",
						zap.Error(err))
				}

				if err := stream.Send(n); err != nil {
					logger.Fatal("error sending message",
						zap.Error(err))
				}
				grpcMessagesSent.WithLabelValues("stream").Inc()
				logger.Info("sent")

				time.Sleep(waitDuration)
			}
		}()

		// Send a message to start the back and forth Notes
		if err := stream.Send(&messaging.Message{Sender: "client", Message: "Startup"}); err != nil {
			logger.Fatal("error sending startup message",
				zap.Error(err))
		}

		logger.Info("sent startup message")
		<-waitc
	} else {
		for {
			_, err := client.SendMessage(context.Background(), &messaging.Message{Sender: "client", Message: "A single message from a client"})
			if err != nil {
				logger.Error("error sending message",
					zap.Error(err))
			} else {
				grpcMessagesSent.WithLabelValues("unary").Inc()
				logger.Info("sent")
			}

			time.Sleep(waitDuration)
		}
	}
}