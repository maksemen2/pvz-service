package app

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/maksemen2/pvz-service/internal/delivery/http/routes"
	httpserver "github.com/maksemen2/pvz-service/internal/delivery/http/server"
	"github.com/maksemen2/pvz-service/internal/domain/repositories"
	"github.com/maksemen2/pvz-service/internal/pkg/auth"
	"net"

	"github.com/maksemen2/pvz-service/config"
	grpcserver "github.com/maksemen2/pvz-service/internal/delivery/grpc/server"
	"github.com/maksemen2/pvz-service/internal/pkg/auth/jwt"
	"github.com/maksemen2/pvz-service/internal/pkg/database"
	"github.com/maksemen2/pvz-service/internal/pkg/logger"
	"github.com/maksemen2/pvz-service/internal/pkg/metrics"
	postgresqlrepo "github.com/maksemen2/pvz-service/internal/repository/postgresql"
	"github.com/maksemen2/pvz-service/internal/service"
	"go.uber.org/zap"
)

type Application struct {
	Config       *config.Config
	Logger       *zap.Logger
	Database     *database.PostgresDB
	Repositories *Repositories
	Services     *Services
	TokenManager auth.TokenManager
}

type Repositories struct {
	Product   repositories.IProductRepo
	PVZ       repositories.IPVZRepo
	User      repositories.IUserRepo
	Reception repositories.IReceptionRepo
}

type Services struct {
	Auth      service.AuthService
	Product   service.ProductService
	PVZ       service.PVZService
	Reception service.ReceptionService
}

type Servers struct {
	HTTPServer *httpserver.Server
	GRPCServer *grpcserver.Server
	Metrics    *metrics.Server
}

func Initialize(cfg *config.Config) (*Application, error) {
	log, err := logger.NewZapLogger(cfg.Logging)
	if err != nil {
		return nil, fmt.Errorf("logger initialization failed: %w", err)
	}

	db, err := database.NewPostgresDB(cfg.Database, log)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	repos := InitializeRepositories(db, log)
	tokenManager := jwt.NewJWTManager(cfg.Auth)

	services := InitializeServices(repos, log, tokenManager)

	return &Application{
		Config:       cfg,
		Logger:       log,
		Database:     db,
		Repositories: repos,
		Services:     services,
		TokenManager: tokenManager,
	}, nil
}

func (a *Application) StartServers() (*Servers, error) {
	router := a.BuildRouter()
	httpServer := httpserver.New(a.Logger, router, a.Config.HTTP)

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", a.Config.GRPC.Port))
	if err != nil {
		return nil, fmt.Errorf("gRPC listener failed: %w", err)
	}

	grpcServer := grpcserver.New(a.Logger, a.Services.PVZ)
	metricsServer := metrics.NewServer(a.Logger, a.Config.Metrics)

	go httpServer.Start()
	go metricsServer.Start()
	go grpcServer.Start(grpcListener)

	return &Servers{
		HTTPServer: httpServer,
		GRPCServer: grpcServer,
		Metrics:    metricsServer,
	}, nil
}

func (a *Application) BuildRouter() *gin.Engine {
	router := routes.New(a.Services.Auth, a.Services.Product, a.Services.PVZ, a.Services.Reception, a.Logger, a.TokenManager, a.Config.HTTP)
	return router
}

func (s *Servers) Stop(ctx context.Context) {
	s.HTTPServer.Stop(ctx)
	s.Metrics.Stop(ctx)
	s.GRPCServer.Stop()
}

func InitializeRepositories(db *database.PostgresDB, log *zap.Logger) *Repositories {
	return &Repositories{
		Product:   postgresqlrepo.NewPostgresqlProductRepository(db, log),
		PVZ:       postgresqlrepo.NewPostgresqlPVZRepository(db, log),
		User:      postgresqlrepo.NewPostgresqlUserRepository(db, log),
		Reception: postgresqlrepo.NewPostgresqlReceptionRepository(db, log),
	}
}

func InitializeServices(repos *Repositories, log *zap.Logger, tokenManager auth.TokenManager) *Services {
	return &Services{
		Auth:      service.NewAuthService(log, repos.User, tokenManager),
		Product:   service.NewProductService(log, repos.Product),
		PVZ:       service.NewPVZService(log, repos.PVZ),
		Reception: service.NewReceptionService(log, repos.Reception),
	}
}
