package main

import (
	"context"
	"flag"
	pb "github.com/jwenz723/grpcdemo/notestream"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
)

var (
	serverAddrFlag = flag.String("server_addr", "", "The server address in the format of host:port")
)

func main() {
	flag.Parse()

	serverAddr := "localhost:8080"

	if *serverAddrFlag != "" {
		serverAddr = *serverAddrFlag
	} else if serverAddrEnv, ok := os.LookupEnv("SERVER_ADDR"); ok {
		serverAddr = serverAddrEnv
	}

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
	go func() {
		for i := 0; i < 5; i++ {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}

			log.Printf("%s: %s\n", in.Sender, in.Message)
			if err := stream.Send(&pb.Note{Sender: "client", Message: "A message from a client"}); err != nil {
				log.Fatalf("Failed to send a note: %v", err)
			}
		}
	}()

	// Send a message to start the back and forth Notes
	if err := stream.Send(&pb.Note{Sender: "client", Message: "Startup"}); err != nil {
		log.Fatalf("Failed to send a note: %v", err)
	}

	<-waitc
}
