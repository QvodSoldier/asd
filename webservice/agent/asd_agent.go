package main

import (
	pb "ggstudy/asd/webservice/agent/grpc"
	"ggstudy/asd/webservice/agent/server"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":12580")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterDebugServer(s, &server.Server{})
	s.Serve(lis)
}
