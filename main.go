package main

import (
	"os"
	"fmt"
	"log"
	"net"
	"context"
	pb "golden-gate/grpc"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedGoldenGateServer
}

func (s *server) RefreshAPIKeyData(context.Context, *pb.RefreshAPIKeyRequest) (*pb.RefreshAPIKeyReply, error) {
	return &pb.RefreshAPIKeyReply{
		CanBeUsed: true,
	}, nil
}

func gRPCServer(port string) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterGoldenGateServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func getPort() string {
   port := os.Getenv("GOLDEN_GATE_LISTEN_PORT")
   if port == "" {
       panic("GOLDEN_GATE_LISTEN_PORT not set.")
   }

   return port
}

func main() {
	port := getPort()
	gRPCServer(port)
}