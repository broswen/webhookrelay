package grpc

import (
	apiv1 "github.com/broswen/webhookrelay/pkg/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// NewGRPCServer creates a new gRPC server with all necessary services registered
func NewGRPCServer(webhookServer *Server) *grpc.Server {
	// Create gRPC server
	s := grpc.NewServer()

	// Register webhook service
	apiv1.RegisterWebhookRelayServiceServer(s, webhookServer)

	// Register reflection service for easier debugging and testing
	reflection.Register(s)

	// Register health check service
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(s, healthServer)
	
	// Set the service as serving
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("webhookrelay.api.v1.WebhookRelayService", healthpb.HealthCheckResponse_SERVING)

	return s
}
