package postgresqlrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/maksemen2/pvz-service/internal/domain/repositories"

	"github.com/maksemen2/pvz-service/internal/domain/models"
	"github.com/maksemen2/pvz-service/internal/pkg/database"
	databaseerrors "github.com/maksemen2/pvz-service/internal/repository/errors"
	"go.uber.org/zap"
)

// postgresqlUserRepository реализует интерфейс repositories.IUserRepo
// для работы с пользователями в PostgreSQL.
type postgresqlUserRepository struct {
	logger *zap.Logger
	db     *database.PostgresDB
}

// NewPostgresqlUserRepository создает новый экземпляр postgresqlUserRepository.
func NewPostgresqlUserRepository(db *database.PostgresDB, logger *zap.Logger) repositories.IUserRepo {
	return &postgresqlUserRepository{
		logger: logger,
		db:     db,
	}
}

// userRow - представление пользователя в базе данных.
type userRow struct {
	ID           uuid.UUID `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	Role         string    `db:"role"`
}

// toModel производит маппинг из представления пользователя в базе данных в доменную модель.
func (r *postgresqlUserRepository) toModel(row userRow) *models.User {
	return &models.User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		Role:         models.RoleType(row.Role),
	}
}

// toRow производит маппинг из доменной модели в представление пользователя в базе данных.
func (r *postgresqlUserRepository) toRow(user *models.User) *userRow {
	return &userRow{
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Role.String(),
	}
}

// Create создает нового пользователя в postgresql.
// Возвращает ошибку, если не удалось создать пользователя или если пользователь с таким email уже существует.
func (r *postgresqlUserRepository) Create(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (id, email, password_hash, role) VALUES (:id, :email, :password_hash, :role)`

	_, err := r.db.NamedExecContext(ctx, query, r.toRow(user))

	if err != nil {
		if database.IsPGError(err, database.PGUniqueViolationCode) {
			return databaseerrors.ErrUniqueViolation
		}

		r.logger.Error("failed to create user", zap.String("email", user.Email), zap.Error(err))

		return fmt.Errorf("%w: %v", databaseerrors.ErrUnexpected, err)
	}

	return nil
}

// GetByEmail получает пользователя по email из postgresql.
// Возвращает ошибку, если пользователь не найден или если возникла непредвиденная ошибка базы данных.
func (r *postgresqlUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, email, password_hash, role FROM users WHERE email = $1`

	var uRow userRow

	err := r.db.GetContext(ctx, &uRow, query, email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, databaseerrors.ErrNoRows
		}

		r.logger.Error("failed to get user by email", zap.String("email", email), zap.Error(err))

		return nil, fmt.Errorf("%w: %v", databaseerrors.ErrUnexpected, err)
	}

	return r.toModel(uRow), nil
}
