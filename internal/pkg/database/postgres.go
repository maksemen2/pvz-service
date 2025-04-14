package database

import (
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/maksemen2/pvz-service/config"
	"go.uber.org/zap"
)

// Коды ошибок PostgreSQL.
// См. https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	PGUniqueViolationCode = "23505" // Уникальное ограничение нарушено
)

// PostgresDB - структура для работы с PostgreSQL.
type PostgresDB struct {
	*sqlx.DB
	logger *zap.Logger
}

// NewPostgresDB - создает новое подключение к PostgreSQL.
func NewPostgresDB(config config.DatabaseConfig, logger *zap.Logger) (*PostgresDB, error) {
	db, err := sqlx.Connect("postgres", config.DSN())
	if err != nil {
		logger.Error("failed to connect to postgres", zap.Error(err))
		return nil, err
	}

	logger.Info("connected to postgres")

	if err := db.Ping(); err != nil {
		logger.Error("failed to ping postgres", zap.Error(err))
		return nil, err
	}

	db.SetMaxOpenConns(config.MaxOpenConnections)
	db.SetMaxIdleConns(config.MaxIdleConnections)

	return &PostgresDB{
		DB:     db,
		logger: logger,
	}, nil
}

// Close закрывает подключение к PostgreSQL.
func (d *PostgresDB) Close() error {
	d.logger.Info("closing postgres connection")

	if err := d.DB.Close(); err != nil {
		d.logger.Error("failed to close postgres connection", zap.Error(err))
	}

	d.logger.Info("postgres connection closed")

	return nil
}

// IsPGError - проверяет, является ли ошибка
// ошибкой PostgreSQL с указанным кодом.
func IsPGError(err error, code string) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == pq.ErrorCode(code)
	}

	return false
}

// TxRollback - хелпер для отката транзакции.
// Предполагается, что эта функция должна использоваться в defer
// Если транзакция уже завершена, то ошибка будет игнорироваться
func TxRollback(tx *sqlx.Tx, logger *zap.Logger) {
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		logger.Error("failed to rollback transaction", zap.Error(err))
	}
}
