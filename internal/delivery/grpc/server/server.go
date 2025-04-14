package grpcserver

import (
	grpchandlers "github.com/maksemen2/pvz-service/internal/delivery/grpc/handlers"
	"github.com/maksemen2/pvz-service/internal/delivery/grpc/pvz_v1"
	"github.com/maksemen2/pvz-service/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

// Server - структура для gRPC сервера.
type Server struct {
	server *grpc.Server
	logger *zap.Logger
}

func New(logger *zap.Logger, pvzService service.PVZService) *Server {
	srv := grpc.NewServer()

	pvz_v1.RegisterPVZServiceServer(srv, grpchandlers.NewPVZServer(pvzService))

	return &Server{
		server: srv,
		logger: logger,
	}
}

// Start - запускает gRPC сервер на указанном слушателе.
// Принимает слушатель net.Listener.
func (s *Server) Start(lis net.Listener) {
	s.logger.Info("Starting gRPC server", zap.String("addr", lis.Addr().String()))

	if err := s.server.Serve(lis); err != nil {
		s.logger.Error("Failed to serve", zap.Error(err))
	}
}

// Stop - останавливает gRPC сервер.
func (s *Server) Stop() {
	s.server.GracefulStop()
}
