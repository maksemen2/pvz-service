//go:build integration
// +build integration

package database_test

import (
	"fmt"
	"github.com/maksemen2/pvz-service/internal/pkg/database"
	"github.com/maksemen2/pvz-service/internal/pkg/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
)

func TestOpenClose(t *testing.T) {
	cfg, cleanup := testhelpers.SetupPostgresContainer(t)
	defer cleanup()

	logger := zap.NewNop()
	db, err := database.NewPostgresDB(cfg, logger)
	require.NoError(t, err)

	err = db.Ping()
	assert.NoError(t, err)

	err = db.Close()
	assert.NoError(t, err)
}

func TestIsPGError(t *testing.T) {
	cfg, cleanup := testhelpers.SetupPostgresContainer(t)
	defer cleanup()

	logger := zap.NewNop()
	db, err := database.NewPostgresDB(cfg, logger)
	require.NoError(t, err)

	defer db.Close()

	_, err = db.Exec(`
        CREATE TABLE test_table (
            id SERIAL PRIMARY KEY,
            name VARCHAR UNIQUE
        );
    `)
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO test_table (name) VALUES ('test');")
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO test_table (name) VALUES ('test');")
	require.Error(t, err)

	assert.True(t, database.IsPGError(err, database.PGUniqueViolationCode))

	nonPgErr := fmt.Errorf("ordinary error")
	assert.False(t, database.IsPGError(nonPgErr, database.PGUniqueViolationCode))
}
