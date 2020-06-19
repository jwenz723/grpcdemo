# grpcNoteStream
grpc demo of sending notes between a server and client written in Go. With both the client and server running you will see messages being continuously passed back and forth between the server and the client.

# Running Server and Client using Skaffold

## In 'Dev' Mode

Dev mode will automatically update the deployed resources any time a new change is saved:

`skaffold dev`

## In 'Run' Mode

Run mode will start deploy the resources once.

`skaffold run`

To view the logs for the deployed resources:

`skaffold run --tail`

To have local port-forwarding setup to the deployed resources:

`skaffold run --port-forward`

# Running Server
Build the code in the Server directory then run it

# Running Client
Build the code in the Client directory then run it

# How to Modify Proto

1. modify messaging.proto file
2. Generate go code from .proto:
`protoc messaging.proto --go_out=plugins=grpc:messaging`