package grpc

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/gongahkia/kite/internal/grpc/services"
	"github.com/gongahkia/kite/internal/observability"
	"github.com/gongahkia/kite/internal/storage"
	"github.com/gongahkia/kite/internal/queue"
	pb "github.com/gongahkia/kite/api/proto"
)

// Server represents the gRPC server
type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	logger     *observability.Logger
	storage    storage.Storage
	queue      queue.Queue
}

// ServerConfig holds gRPC server configuration
type ServerConfig struct {
	Port    int
	Storage storage.Storage
	Queue   queue.Queue
	Logger  *observability.Logger
}

// NewServer creates a new gRPC server
func NewServer(config *ServerConfig) (*Server, error) {
	// Create listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	// Create gRPC server with interceptors
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor(config.Logger),
			recoveryInterceptor(config.Logger),
			metricsInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			streamLoggingInterceptor(config.Logger),
			streamRecoveryInterceptor(config.Logger),
		),
	)

	server := &Server{
		grpcServer: grpcServer,
		listener:   lis,
		logger:     config.Logger,
		storage:    config.Storage,
		queue:      config.Queue,
	}

	// Register services
	server.registerServices()

	// Enable reflection for tools like grpcurl
	reflection.Register(grpcServer)

	return server, nil
}

// registerServices registers all gRPC services
func (s *Server) registerServices() {
	// Register ScraperService
	scraperSvc := services.NewScraperService(s.storage, s.queue, s.logger)
	pb.RegisterScraperServiceServer(s.grpcServer, scraperSvc)

	// Register SearchService
	searchSvc := services.NewSearchService(s.storage, s.logger)
	pb.RegisterSearchServiceServer(s.grpcServer, searchSvc)

	s.logger.Info("gRPC services registered")
}

// Start starts the gRPC server
func (s *Server) Start() error {
	s.logger.Infof("Starting gRPC server on %s", s.listener.Addr().String())
	return s.grpcServer.Serve(s.listener)
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	s.logger.Info("Stopping gRPC server...")
	s.grpcServer.GracefulStop()
}

// GetListener returns the server's listener
func (s *Server) GetListener() net.Listener {
	return s.listener
}
