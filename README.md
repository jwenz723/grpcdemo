# grpcNoteStream
grpc demo of sending notes between a server and client written in Go. With both the client and server running you will see messages being continuously passed back and forth between the server and the client.

# Running Server
Build the code in the Server directory then run it

# Running Client
Build the code in the Client directory then run it

# How to Modify Proto

1. modify messaging.proto file
2. Generate go code from .proto:
`protoc messaging.proto --go_out=plugins=grpc:messaging`