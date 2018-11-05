package main

import (
	pb "github.com/jwenz723/grpcNoteStream/NoteStream"
	"io"
	"google.golang.org/grpc"
	"fmt"
	"net"
	"flag"
	"log"
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

// StreamResponse implements helloworld.GreeterServer
func (s *server) StreamNotes(stream pb.NoteStream_StreamNotesServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		log.Printf("%s: %s\n", in.Sender, in.Message)
		if err := stream.Send(&pb.Note{Sender: "Server", Message:"Do Bad"}); err != nil {
			return err
		}
	}
}

func newServer() *server {
	s := &server{}
	return s
}

var (
	port       = flag.Int("port", 10000, "The server port")
)

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("starting grpc server on port %d\n", *port)
	grpcServer := grpc.NewServer()
	pb.RegisterNoteStreamServer(grpcServer, newServer())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to start grpc server: %v", err)
	}
}
