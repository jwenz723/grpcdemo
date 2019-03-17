package main

import (
	"flag"
	"fmt"
	pb "github.com/jwenz723/grpcdemo/notestream"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
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
		if err := stream.Send(&pb.Note{Sender: "server", Message: "A message from the server"}); err != nil {
			return err
		}
	}
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

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
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
