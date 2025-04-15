package postgresqlrepo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	domainerrors "github.com/maksemen2/pvz-service/internal/domain/errors"
	"github.com/maksemen2/pvz-service/internal/domain/models"
	"github.com/maksemen2/pvz-service/internal/domain/repositories"
	"github.com/maksemen2/pvz-service/internal/pkg/database"
	databaseerrors "github.com/maksemen2/pvz-service/internal/repository/errors"
	"go.uber.org/zap"
)

// postgresqlReceptionRepository реализует интерфейс repositories.IReceptionRepo
type postgresqlReceptionRepository struct {
	db     *database.PostgresDB
	logger *zap.Logger
}

// NewPostgresqlReceptionRepository создает новый экземпляр postgresqlReceptionRepository.
func NewPostgresqlReceptionRepository(db *database.PostgresDB, logger *zap.Logger) repositories.IReceptionRepo {
	return &postgresqlReceptionRepository{
		db:     db,
		logger: logger,
	}
}

// receptionRow - представляет собой строку из таблицы receptions в базе данных.
type receptionRow struct {
	ID       uuid.UUID `db:"id"`
	DateTime time.Time `db:"date_time"`
	PVZID    uuid.UUID `db:"pvz_id"`
	Status   string    `db:"status"`
}

// toModel производит маппинг из строки таблицы receptions в доменную модель.
func (r *postgresqlReceptionRepository) toModel(row receptionRow) *models.Reception {
	return &models.Reception{
		ID:       row.ID,
		DateTime: row.DateTime,
		PVZID:    row.PVZID,
		Status:   models.ReceptionStatus(row.Status),
	}
}

// toRow производит маппинг из доменной модели в строку таблицы receptions.
func (r *postgresqlReceptionRepository) toRow(reception *models.Reception) *receptionRow {
	return &receptionRow{
		ID:       reception.ID,
		DateTime: reception.DateTime,
		PVZID:    reception.PVZID,
		Status:   reception.Status.String(),
	}
}

// CreateIfNoOpen создает новую приемку, если в ПВЗ нет открытых приемок.
// В транзакции проверяет наличие ПВЗ и открытых приемок.
// Если открытая приёмка уже существует, возвращает ошибку.
func (r *postgresqlReceptionRepository) CreateIfNoOpen(ctx context.Context, reception *models.Reception) error {
	// Можно было бы использовать меньше запросов, но такой подход позволяет
	// 1) Возвращать более детализованные ошибки
	// 2) Улучшить читаемость кода
	// 3) Улучшить производительность в определенных сценариях (например, когда на вход сразу подается айди не существующего ПВЗ)
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		r.logger.Error("failed to begin transaction", zap.Error(err))
		return databaseerrors.ErrUnexpected
	}
	defer database.TxRollback(tx, r.logger)

	var pvzExists bool

	err = tx.GetContext(ctx, &pvzExists,
		"SELECT EXISTS(SELECT 1 FROM pvzs WHERE id = $1)",
		reception.PVZID,
	)
	if err != nil {
		r.logger.Error("error checking PVZ existence", zap.Error(err))
		return databaseerrors.ErrUnexpected
	}

	if !pvzExists {
		return databaseerrors.ErrNoRows
	}

	var openReceptionExists bool

	err = tx.GetContext(ctx, &openReceptionExists,
		"SELECT EXISTS(SELECT 1 FROM receptions WHERE pvz_id = $1 AND status = 'in_progress')",
		reception.PVZID,
	)
	if err != nil {
		r.logger.Error("error checking open receptions", zap.Error(err))
		return databaseerrors.ErrUnexpected
	}

	if openReceptionExists {
		return domainerrors.ErrOpenReceptionExists
	}

	_, err = tx.NamedExecContext(ctx,
		`INSERT INTO receptions (id, date_time, pvz_id, status)
         VALUES (:id, :date_time, :pvz_id, :status)`,
		r.toRow(reception),
	)
	if err != nil {
		r.logger.Error("error inserting reception", zap.Error(err))
		return databaseerrors.ErrUnexpected
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("failed to commit transaction", zap.Error(err))
		return databaseerrors.ErrUnexpected
	}

	return nil
}

// CloseLast закрывает последнюю открывшуюся приемку в ПВЗ.
// Если приемка не найдена, возвращает ошибку.
func (r *postgresqlReceptionRepository) CloseLast(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {

	var receptionRow receptionRow

	err := r.db.GetContext(ctx, &receptionRow, `
        UPDATE receptions 
        SET status = 'close' 
        WHERE pvz_id = $1
		AND status = 'in_progress'
        RETURNING id, date_time, pvz_id, status`,
		pvzID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrNoOpenReceptions
		}

		r.logger.Error("failed to close reception", zap.Error(err))

		return nil, databaseerrors.ErrUnexpected
	}

	return r.toModel(receptionRow), nil
}
