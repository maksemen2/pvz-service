package metrics

import (
	"context"
	"github.com/maksemen2/pvz-service/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
)

// Server - структура, представляющая сервер метрик.
type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
}

// NewServer принимает логгер и конфиг метрик.
// Создает ServeMux и регистрирует обработчик для метрик.
// Возвращает указатель на Server.
func NewServer(logger *zap.Logger, cfg config.MetricsConfig) *Server {
	mux := http.NewServeMux()
	mux.Handle(cfg.Path, promhttp.HandlerFor(Registry, promhttp.HandlerOpts{}))
	srv := &http.Server{
		Addr:    cfg.GetAddr(),
		Handler: mux,
	}

	return &Server{
		httpServer: srv,
		logger:     logger,
	}
}

// Start запускает сервер метрик.
// Если происходит ошибка - она логируется и программа завершается.
func (s *Server) Start() {
	s.logger.Info("Starting metrics server", zap.String("addr", s.httpServer.Addr))

	if err := s.httpServer.ListenAndServe(); err != nil {
		s.logger.Fatal("Metrics server error", zap.Error(err))
	}
}

// Stop остановливает сервер метрик.
// Принимает контекст для корректного завершения работы.
func (s *Server) Stop(ctx context.Context) {
	s.logger.Info("Stopping metrics server")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Error stopping metrics server", zap.Error(err))
	}

	s.logger.Info("Metrics server stopped")
}
