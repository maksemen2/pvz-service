package config

import "fmt"

// Config объединяет в себе все другие
// конфиги для отдельных сервисов
type Config struct {
	HTTP     HTTPConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Metrics  MetricsConfig
	Logging  LoggingConfig
	GRPC     GRPCConfig
}

// HTTPConfig содержит конфигурацию
// сервера
type HTTPConfig struct {
	Host string `env:"HTTP_HOST" env-default:"0.0.0.0"`
	Port int    `env:"HTTP_PORT" env-default:"8080"`
	Env  string `env:"ENV" env-default:"dev"` // dev или prod.
}

// GetAddr возвращает адрес сервера
func (c *HTTPConfig) GetAddr() string {
	if c.Host == "localhost" {
		return fmt.Sprintf(":%d", c.Port)
	}

	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// DatabaseConfig содержит конфигурацию
// подключения к БД. Дефолтные значения
// указаны для Postgres
type DatabaseConfig struct {
	Host               string `env:"DB_HOST" env-default:"localhost"`
	Port               int    `env:"DB_PORT" env-default:"5432"`
	User               string `env:"DB_USER" env-default:"postgres"`
	Password           string `env:"DB_PASSWORD" env-default:"postgres"`
	DBName             string `env:"DB_NAME" env-default:"pvz_service"`
	SSLMode            string `env:"DB_SSLMODE" env-default:"disable"`
	MaxOpenConnections int    `env:"DB_MAX_OPEN_CONNS" env-default:"25"`
	MaxIdleConnections int    `env:"DB_MAX_IDLE_CONNS" env-default:"5"`
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// AuthConfig содержит конфигурацию
// Для JWT аутентификации
type AuthConfig struct {
	JWTSecret              string `env:"JWT_SECRET" env-required:"true"`
	TokenExpirationSeconds int    `env:"TOKEN_EXPIRATION" env-default:"3600"` // Время в секундах
}

// MetricsConfig содержит конфигурацию для
// Сбора метрик из Prometheus. Дефолтные значения
// взяты из описания задания.
type MetricsConfig struct {
	Port int    `env:"METRICS_PORT" env-default:"9000"`
	Path string `env:"METRICS_PATH" env-default:"/metrics"`
}

func (m *MetricsConfig) GetAddr() string {
	return fmt.Sprintf(":%d", m.Port)
}

// LoggingConfig содержит конфигурацию для
// логгера.
type LoggingConfig struct {
	Level string `env:"LOG_LEVEL" env-default:"info"` // debug, info, warn, error или silent
}

// GRPCConfig содержит конфигурацию для gRPC сервера.
// Дефолтный порт взят из описания задания.
type GRPCConfig struct {
	Port int `env:"GRPC_PORT" env-default:"3000"`
}
