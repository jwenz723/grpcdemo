package main

import (
	"context"
	"flag"
	pb "github.com/jwenz723/grpcdemo/messaging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	serverAddrFlag   = flag.String("server_addr", "", "The server address in the format of host:port")
	useStreaming     = flag.Bool("use_streaming", false, "Setting this will use grpc streaming instead of repeated single messages")
	waitNanos        = flag.Int("wait_nanos", 500, "The number of nanoseconds to wait before sending messages (this applies to both single and stream messages)")
	grpcMessagesSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "grpc_messages_sent",
			Help:        "The total number of grpc messages sent",
			ConstLabels: prometheus.Labels{"from": "client"},
		},
		[]string{"method"},
	)
	waitDuration time.Duration
)

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(grpcMessagesSent)
}

func main() {
	flag.Parse()

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
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewMessagingServiceClient(conn)

	if *useStreaming {
		waitc := make(chan struct{})
		stream, err := client.StreamMessages(context.Background())
		if err != nil {
			log.Fatalf("error opening stream: %v", err)
		}
		defer stream.CloseSend()

		n := &pb.Message{Sender: "client", Message: "A streamed message from a client"}
		go func() {
			for {
				m, err := stream.Recv()
				if err == io.EOF {
					// read done.
					close(waitc)
					return
				}
				if err != nil {
					log.Fatalf("error receiving message : %v", err)
				}

				if err := stream.Send(n); err != nil {
					log.Fatalf("error sending message: %v", err)
				}

				handleReceivedMessage(m, "stream")

				time.Sleep(waitDuration)
			}
		}()

		// Send a message to start the back and forth Notes
		if err := stream.Send(&pb.Message{Sender: "client", Message: "Startup"}); err != nil {
			log.Fatalf("error sending startup message: %v", err)
		}

		log.Printf("sent startup message\n")
		<-waitc
	} else {
		for {
			m, err := client.SendMessage(context.Background(), &pb.Message{Sender: "client", Message: "A single message from a client"})
			if err != nil {
				log.Printf("error sending message: %v\n", err)
			}

			handleReceivedMessage(m, "single")

			time.Sleep(waitDuration)
		}
	}
}

func handleReceivedMessage(m *pb.Message, receiveType string) {
	grpcMessagesSent.WithLabelValues(receiveType).Inc()
	log.Printf("received from server: %s - %s, waiting %d millis\n", m.Sender, m.Message, *waitNanos)
}
