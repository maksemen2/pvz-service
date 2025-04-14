package testhelpers

import (
	"context"
	"fmt"
	"github.com/maksemen2/pvz-service/internal/pkg/database"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/maksemen2/pvz-service/config"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupPostgresContainer создает контейнер PostgreSQL для тестов.
// Возвращает тестовую конфигурацию базы данных и функцию для очистки контейнера.
func SetupPostgresContainer(t *testing.T) (config.DatabaseConfig, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:13",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "test",
		},
		WaitingFor: wait.ForSQL(
			"5432/tcp",
			"postgres",
			func(host string, port nat.Port) string {
				return fmt.Sprintf(
					"postgres://test:test@%s:%s/test?sslmode=disable",
					host, port.Port(),
				)
			},
		).WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	cfg := config.DatabaseConfig{
		Host:               "127.0.0.1",
		Port:               port.Int(),
		User:               "test",
		Password:           "test",
		DBName:             "test",
		SSLMode:            "disable",
		MaxOpenConnections: 5,
		MaxIdleConnections: 2,
	}

	return cfg, func() {
		if container != nil {
			_ = container.Terminate(ctx)
		}
	}
}

// CreateTestDB создает тестовую базу данных с необходимыми таблицами.
// Возвращает функцию для очистки базы данных после тестов.
func CreateTestDB(db *database.PostgresDB) (func(), error) {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS pvzs (
	id UUID PRIMARY KEY,
	registration_date TIMESTAMP NOT NULL,
	city VARCHAR(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS receptions (
	id UUID PRIMARY KEY,
	pvz_id UUID NOT NULL REFERENCES pvzs(id),
	date_time TIMESTAMP NOT NULL,
	status VARCHAR(20) NOT NULL
);

CREATE TABLE IF NOT EXISTS products (
	id UUID PRIMARY KEY,
	date_time TIMESTAMP NOT NULL,
	type VARCHAR(50) NOT NULL,
	reception_id UUID NOT NULL REFERENCES receptions(id)
);

CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY,
	email VARCHAR(255) UNIQUE NOT NULL,
	password_hash VARCHAR(255) NOT NULL,
	role VARCHAR(50) NOT NULL
);
`)

	cleanup := func() {
		_, _ = db.Exec("DROP TABLE IF EXISTS products")
		_, _ = db.Exec("DROP TABLE IF EXISTS receptions")
		_, _ = db.Exec("DROP TABLE IF EXISTS pvzs")
		_, _ = db.Exec("DROP TABLE IF EXISTS users")
	}

	return cleanup, err
}
