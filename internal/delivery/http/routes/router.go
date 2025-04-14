package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/maksemen2/pvz-service/config"
	httphandlers "github.com/maksemen2/pvz-service/internal/delivery/http/handlers"
	"github.com/maksemen2/pvz-service/internal/pkg/auth"
	l "github.com/maksemen2/pvz-service/internal/pkg/logger"
	"github.com/maksemen2/pvz-service/internal/pkg/metrics"
	"github.com/maksemen2/pvz-service/internal/service"
	"go.uber.org/zap"
)

// New настраивает роутинг приложения и устанавливает мидлвари.
// Возвращает инстанс gin.Engine
func New(authService service.AuthService, productService service.ProductService, pvzService service.PVZService, receptionService service.ReceptionService, logger *zap.Logger, tokenManager auth.TokenManager, config config.HTTPConfig) *gin.Engine {
	router := gin.New()

	if config.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router.Use(gin.Recovery(), metrics.NewGinMiddleware(), l.NewMiddleware(logger))

	public := router.Group("")

	authHandler := httphandlers.NewAuthHandler(logger, authService)

	authHandler.RegisterRoutes(public)

	protected := router.Group("")

	protected.Use(auth.NewGinMiddleware(logger, tokenManager))

	productHandler := httphandlers.NewProductHandler(logger, productService)

	productHandler.RegisterRoutes(protected)

	pvzHandler := httphandlers.NewPVZHandler(logger, pvzService)

	pvzHandler.RegisterRoutes(protected)

	receptionHandler := httphandlers.NewReceptionHandler(logger, receptionService)

	receptionHandler.RegisterRoutes(protected)

	return router
}
