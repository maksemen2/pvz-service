package httpserver

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/maksemen2/pvz-service/config"
	"go.uber.org/zap"
	"net/http"
)

// Server - структура HTTP-сервера для обработки API запросов.
type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
}

func New(logger *zap.Logger, router *gin.Engine, cfg config.HTTPConfig) *Server {
	srv := &http.Server{
		Addr:    cfg.GetAddr(),
		Handler: router,
	}

	return &Server{
		httpServer: srv,
		logger:     logger,
	}
}

// Start запускает HTTP сервер
func (s *Server) Start() {
	s.logger.Info("Starting HTTP server", zap.String("addr", s.httpServer.Addr))

	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Fatal("HTTP server error", zap.Error(err))
	}
}

// Stop останавливает HTTP сервер
// GracefulShutdown должен быть обеспечен снаружи
func (s *Server) Stop(ctx context.Context) {
	s.logger.Info("Stopping HTTP server")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Error stopping HTTP server", zap.Error(err))
	}

	s.logger.Info("HTTP server stopped")
}
