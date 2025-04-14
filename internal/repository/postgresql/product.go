package postgresqlrepo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	domainerrors "github.com/maksemen2/pvz-service/internal/domain/errors"
	"github.com/maksemen2/pvz-service/internal/domain/models"
	"github.com/maksemen2/pvz-service/internal/domain/repositories"
	"github.com/maksemen2/pvz-service/internal/pkg/database"
	databaseerrors "github.com/maksemen2/pvz-service/internal/repository/errors"
	"go.uber.org/zap"
)

// postgresqlProductRepository - структура репозитория для работы с товарами в PostgreSQL.
// Реализует интерфейс repositories.IProductRepo
type postgresqlProductRepository struct {
	db     *database.PostgresDB
	logger *zap.Logger
}

// NewPostgresqlProductRepository - конструктор для создания нового экземпляра postgresqlProductRepository.
// Принимает базу данных и логгер.
func NewPostgresqlProductRepository(db *database.PostgresDB, logger *zap.Logger) repositories.IProductRepo {
	return &postgresqlProductRepository{
		db:     db,
		logger: logger,
	}
}

// productRow - структура для представления строки товара в базе данных.
type productRow struct {
	ID          uuid.UUID `db:"id"`
	DateTime    time.Time `db:"date_time"`
	Type        string    `db:"type"`
	ReceptionID uuid.UUID `db:"reception_id"`
}

// toModel - преобразует строку базы данных в доменную модель товара.
func (r *postgresqlProductRepository) toModel(row productRow) *models.Product {
	return &models.Product{
		ID:          row.ID,
		DateTime:    row.DateTime,
		Type:        models.ProductType(row.Type),
		ReceptionID: row.ReceptionID,
	}
}

// getOpenReceptionID - хелпер для получения ID открытой приёмки в ПВЗ в транзакции.
// Принимает контекст, транзакцию, ID ПВЗ и указатель на ID приёмки.
// Возвращает ошибку, если не удалось найти открытые приёмки или произошла ошибка базы данных.
func (r *postgresqlProductRepository) getOpenReceptionID(ctx context.Context, tx *sqlx.Tx, pvzID uuid.UUID, target *uuid.UUID) error {
	err := tx.GetContext(ctx, target, `
        SELECT id 
        FROM receptions 
        WHERE 
            pvz_id = $1 AND 
            status = 'in_progress'
        ORDER BY date_time DESC 
        LIMIT 1`,
		pvzID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainerrors.ErrNoOpenReceptions
		}

		r.logger.Error("Failed to find open reception", zap.Error(err))

		return databaseerrors.ErrUnexpected
	}

	return nil
}

// Create - создает новый товар в базе данных.
// В транзакции проверяет, существует ли для указанного ПВЗ открытая приёмка, если не существует - возвращает ошибку.
// Если существует - создает новый товар в этой приёмке.
// Возвращает созданный товар или ошибку, если она возникла.
func (r *postgresqlProductRepository) Create(ctx context.Context, product *models.AddProduct) (*models.Product, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to start transaction", zap.Error(err))
		return nil, databaseerrors.ErrUnexpected
	}
	defer database.TxRollback(tx, r.logger)

	var receptionID uuid.UUID
	// Ищем айди последней открытой приемки
	err = r.getOpenReceptionID(ctx, tx, product.PVZID, &receptionID)
	if err != nil {
		return nil, err
	}

	query := `
        INSERT INTO products (id, date_time, type, reception_id)
        VALUES (:id, :date_time, :type, :reception_id)
    `

	row := productRow{
		ID:          product.ID,
		DateTime:    product.DateTime,
		Type:        product.Type.String(),
		ReceptionID: receptionID,
	}

	_, err = tx.NamedExecContext(ctx, query, &row)

	if err != nil {
		r.logger.Error("Failed to create product",
			zap.Error(err),
			zap.String("receptionID", receptionID.String()),
		)

		return nil, databaseerrors.ErrUnexpected
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("Failed to commit transaction", zap.Error(err))
		return nil, databaseerrors.ErrUnexpected
	}

	return r.toModel(row), nil
}

// DeleteLast - удаляет последний товар из открытой приёмки в указанном ПВЗ.
// Проверяет, есть ли открытая приёмка в ПВЗ и получает её айди.
// Если открытая приёмка найдена - удаляет последний товар из неё.
// Если товаров нет или нет открытой приёмки - возвращает ошибку.
func (r *postgresqlProductRepository) DeleteLast(ctx context.Context, pvzID uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		r.logger.Error("Error starting transaction", zap.Error(err))
		return databaseerrors.ErrUnexpected
	}
	defer database.TxRollback(tx, r.logger)

	var receptionID uuid.UUID
	err = r.getOpenReceptionID(ctx, tx, pvzID, &receptionID)

	if err != nil {
		return err
	}

	// Сразу удаляем последний товар в приёмке
	result, err := tx.ExecContext(ctx, `
        DELETE FROM products
        WHERE id = (
            SELECT id FROM products
            WHERE reception_id = $1
            ORDER BY date_time DESC
            LIMIT 1
        )
    `,
		receptionID,
	)
	if err != nil {
		r.logger.Error("Error deleting last product", zap.Error(err))
		return databaseerrors.ErrUnexpected
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Error getting rows affected", zap.Error(err))
		return databaseerrors.ErrUnexpected
	}

	r.logger.Debug("Rows affected after deleting product", zap.Int64("rowsAffected", rowsAffected))

	if rowsAffected == 0 {
		// Если не удалили ни одного товара - значит, их и не было
		return domainerrors.ErrNoProductsInReception
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("Error committing transaction", zap.Error(err))
		return databaseerrors.ErrUnexpected
	}

	return nil
}
