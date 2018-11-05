package main

import (
	"google.golang.org/grpc"
	"flag"
	"log"
	pb "github.com/jwenz723/grpcNoteStream/NoteStream"
	"context"
	"io"
)

var (
	serverAddr         = flag.String("server_addr", "127.0.0.1:10000", "The server address in the format of host:port")
)

func main() {
	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
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
		for {
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
			if err := stream.Send(&pb.Note{Sender: "Client", Message:"Do Good"}); err != nil {
				log.Fatalf("Failed to send a note: %v", err)
			}
		}
	}()

	// Send a message to start the back and forth Notes
	if err := stream.Send(&pb.Note{Sender: "Client", Message:"Startup"}); err != nil {
		log.Fatalf("Failed to send a note: %v", err)
	}

	<-waitc
}
