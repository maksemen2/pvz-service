package testhelpers

import (
	"github.com/maksemen2/pvz-service/config"
	"net/http"
	"net/http/httptest"
	"testing"
)

func SetupTestEnvironment(t *testing.T) (*config.Config, func()) {
	dbCfg, cleanup := SetupPostgresContainer(t)

	httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	return &config.Config{
		Database: dbCfg,
		HTTP:     config.HTTPConfig{Port: 8081, Host: "localhost", Env: "prod"},
		Auth:     config.AuthConfig{JWTSecret: "test-secret", TokenExpirationSeconds: 3600},
		Metrics:  config.MetricsConfig{Port: 9001, Path: "/metrics"},
		Logging:  config.LoggingConfig{Level: "silent"},
		GRPC:     config.GRPCConfig{Port: 3001},
	}, cleanup
}
