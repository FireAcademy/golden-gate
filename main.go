package main

import (
	"os"
	"fmt"
	"log"
	"net"
	"context"
	"google.golang.org/grpc"
	pb "github.com/fireacademy/golden-gate/grpc"
	. "github.com/fireacademy/golden-gate/redis"
	telemetry "github.com/fireacademy/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

type server struct {
	pb.UnimplementedGoldenGateServer
}

func (s *server) RefreshAPIKeyData(ctx context.Context, r *pb.RefreshAPIKeyRequest) (*pb.RefreshAPIKeyReply, error) {
	canBeUsed, origin, err := RefreshAPIKey(ctx, r.APIKey)

	return &pb.RefreshAPIKeyReply{
		CanBeUsed: canBeUsed,
		Origin: origin,
	}, err
}

func (s *server) BillCredits(ctx context.Context, r *pb.BillCreditsRequest) (*pb.EmptyReply, error) {
	err := BillCreditsQuickly(ctx, r.APIKey, r.Credits)
	return &pb.EmptyReply{}, err
}

func gRPCServer(port string) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)
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
	cleanup := telemetry.Initialize()
	defer cleanup(context.Background())

	SetupRedis()
	SetupCheck()

	port := getPort()
	gRPCServer(port)
}