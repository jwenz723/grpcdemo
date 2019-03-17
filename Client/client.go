package main

import (
	"context"
	"flag"
	pb "github.com/jwenz723/grpcdemo/notestream"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"io"
	"log"
	"net/http"
	"os"
)

var (
	serverAddrFlag   = flag.String("server_addr", "", "The server address in the format of host:port")
	grpcMessagesSent = promauto.NewCounter(prometheus.CounterOpts{
		Name: "grpc_messages_sent",
		Help: "The total number of grpc messages sent",
	})
)

func main() {
	flag.Parse()

	serverAddr := "localhost:8080"

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

	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewNoteStreamClient(conn)
	stream, err := client.StreamNotes(context.Background())
	if err != nil {
		log.Fatalf("failed to open stream: %v", err)
	}
	defer stream.CloseSend()

	waitc := make(chan struct{})
	n := &pb.Note{Sender: "client", Message: "A message from a client"}
	go func() {
		for {
			_, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}

			//log.Printf("%s: %s\n", in.Sender, in.Message)
			if err := stream.Send(n); err != nil {
				log.Fatalf("Failed to send a note: %v", err)
			}
			grpcMessagesSent.Inc()
		}
	}()

	// Send a message to start the back and forth Notes
	if err := stream.Send(&pb.Note{Sender: "client", Message: "Startup"}); err != nil {
		log.Fatalf("Failed to send a note: %v", err)
	}

	<-waitc
}
