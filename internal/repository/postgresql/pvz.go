package postgresqlrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/maksemen2/pvz-service/internal/domain/models"
	"github.com/maksemen2/pvz-service/internal/domain/repositories"
	"github.com/maksemen2/pvz-service/internal/pkg/database"
	databaseerrors "github.com/maksemen2/pvz-service/internal/repository/errors"
	"go.uber.org/zap"
)

// postgresqlPVZRepository реализует интерфейс
// repositories.IPVZRepo для работы с ПВЗ в PostgreSQL.
type postgresqlPVZRepository struct {
	logger *zap.Logger
	db     *database.PostgresDB
}

// NewPostgresqlPVZRepository создает новый экземпляр postgresqlPVZRepository.
func NewPostgresqlPVZRepository(db *database.PostgresDB, logger *zap.Logger) repositories.IPVZRepo {
	return &postgresqlPVZRepository{
		logger: logger,
		db:     db,
	}
}

// pvzRow представляет собой строку из таблицы pvzs в базе данных.
type pvzRow struct {
	ID               uuid.UUID `db:"id"`
	RegistrationDate time.Time `db:"registration_date"`
	City             string    `db:"city"`
}

// listedPVZRow представляет собой строку из представления ПВЗ в базе данных,
// которая включает в себя информацию о приемках и товарах.
type listedPVZRow struct {
	ID               uuid.UUID  `db:"id"` // айди пвз
	RegistrationDate time.Time  `db:"registration_date"`
	City             string     `db:"city"`
	ReceptionID      *uuid.UUID `db:"reception_id"`
	ReceptionDate    *time.Time `db:"reception_date"`
	ReceptionStatus  *string    `db:"reception_status"`
	ProductID        *uuid.UUID `db:"product_id"`
	ProductDate      *time.Time `db:"product_date"`
	ProductType      *string    `db:"product_type"`
}

// toModel производит маппинг из представления ПВЗ в базе данных в доменную модель.
func (r *postgresqlPVZRepository) toModel(row pvzRow) *models.PVZ {
	return &models.PVZ{
		ID:               row.ID,
		RegistrationDate: row.RegistrationDate,
		City:             models.CityType(row.City),
	}
}

// toRow производит маппинг из доменной модели в представление ПВЗ в базе данных.
func (r *postgresqlPVZRepository) toRow(pvz *models.PVZ) *pvzRow {
	return &pvzRow{
		ID:               pvz.ID,
		RegistrationDate: pvz.RegistrationDate,
		City:             string(pvz.City),
	}
}

// listedToModel производит маппинг из представления ПВЗ в базе данных в доменную модель с приемками и товарами.
func (r *postgresqlPVZRepository) listedToModel(rows []*listedPVZRow) []*models.PVZWithReceptions {
	pvzMap := make(map[uuid.UUID]*models.PVZWithReceptions)
	receptionMap := make(map[uuid.UUID]*models.ReceptionWithProducts)

	for _, row := range rows {
		if _, exists := pvzMap[row.ID]; !exists {
			pvzMap[row.ID] = &models.PVZWithReceptions{
				PVZ: &models.PVZ{
					ID:               row.ID,
					RegistrationDate: row.RegistrationDate,
					City:             models.CityType(row.City),
				},
				Receptions: []*models.ReceptionWithProducts{},
			}
		}

		if row.ReceptionID != nil {
			if _, exists := receptionMap[*row.ReceptionID]; !exists {
				receptionMap[*row.ReceptionID] = &models.ReceptionWithProducts{
					Reception: &models.Reception{
						ID:       *row.ReceptionID,
						DateTime: *row.ReceptionDate,
						PVZID:    row.ID,
						Status:   models.ReceptionStatus(*row.ReceptionStatus),
					},
					Products: []*models.Product{},
				}
				pvzMap[row.ID].Receptions = append(pvzMap[row.ID].Receptions, receptionMap[*row.ReceptionID])
			}
			// Если присутствует товар, добавляем его в соответствующую приемку
			if row.ProductID != nil {
				receptionMap[*row.ReceptionID].Products = append(receptionMap[*row.ReceptionID].Products, &models.Product{
					ID:          *row.ProductID,
					DateTime:    *row.ProductDate,
					Type:        models.ProductType(*row.ProductType),
					ReceptionID: *row.ReceptionID,
				})
			}
		}
	}

	result := make([]*models.PVZWithReceptions, 0, len(pvzMap))
	for _, pvz := range pvzMap {
		result = append(result, pvz)
	}

	return result
}

// Create создает новую запись о пвз в базе данных.
// Возвращает ошибку если что-то пошло не так или если пвз с указанным айди уже существует.
func (r *postgresqlPVZRepository) Create(ctx context.Context, pvz *models.PVZ) error {
	query := `INSERT INTO pvzs (id, registration_date, city) VALUES (:id, :registration_date, :city)`

	_, err := r.db.NamedExecContext(ctx, query, r.toRow(pvz))

	if err != nil {
		if database.IsPGError(err, database.PGUniqueViolationCode) {
			return databaseerrors.ErrUniqueViolation
		}

		r.logger.Error("failed to create PVZ", zap.Error(err))

		return databaseerrors.ErrUnexpected
	}

	return nil
}

// List выводит список ПВЗ, приемок в них и товаров в приёмках с пагинацией и фильтром по времени. (см. models.PVZFilter).
// Пагинация затрагивает только ПВЗ (влияет на количество ПВЗ в результате)
// Фильтрация по времени затрагивает только приёмки (если фильтр по дате не указан -
// выведутся все ПВЗ даже без приёмок, если указан - только те, в которых есть приёмки, входящие в диапазон)
// Возвращает ошибку, если произошла ошибка при выполнении запроса к базе данных.
func (r *postgresqlPVZRepository) List(ctx context.Context, filter *models.PVZFilter) ([]*models.PVZWithReceptions, error) {
	// Основной запрос с подзапросом для пагинации.
	// Первая %s - это условие для JOIN, вторая %s - это WHERE.
	baseQuery := `
        WITH paginated_pvz AS (
            SELECT id
            FROM pvzs
            ORDER BY registration_date DESC
            LIMIT $1
            OFFSET $2
        )
        SELECT
            p.id,
            p.registration_date,
            p.city,
            r.id as reception_id,
            r.date_time as reception_date,
            r.status as reception_status,
            pr.id as product_id,
            pr.date_time as product_date,
            pr.type as product_type
        FROM paginated_pvz pp
        INNER JOIN pvzs p ON pp.id = p.id
        %s
        %s
        ORDER BY p.registration_date DESC, r.date_time DESC
    `

	var joinClause, whereClause string

	args := []interface{}{
		filter.PageSize,
		(filter.Page - 1) * filter.PageSize,
	}

	if filter.StartDate != nil || filter.EndDate != nil {
		// Если у нас есть фильтр по дате - делаем INNER JOIN, исключить ПВЗ, в которых вообще нет приёмок
		joinClause = `
            INNER JOIN receptions r ON p.id = r.pvz_id
            LEFT JOIN products pr ON r.id = pr.reception_id
        `
		// Фильтруем по дате приёмок
		whereClause = `
            WHERE 
                (r.date_time >= $3 OR $3 IS NULL) AND
                (r.date_time <= $4 OR $4 IS NULL)
        `

		args = append(args, filter.StartDate, filter.EndDate)
	} else {
		// А тут, если фильтра по дате нет - делаем LEFT JOIN, тем самым получая ПВЗ даже без приёмок
		joinClause = `
            LEFT JOIN receptions r ON p.id = r.pvz_id
            LEFT JOIN products pr ON r.id = pr.reception_id
        `
		// И условие по дате нам уже не нужно
		whereClause = ""
	}

	finalQuery := fmt.Sprintf(baseQuery, joinClause, whereClause)

	rows, err := r.db.QueryxContext(ctx, finalQuery, args...)
	if err != nil {
		r.logger.Error("failed to list PVZs", zap.Error(err))
		return nil, databaseerrors.ErrUnexpected
	}
	defer rows.Close()

	var rawRows []*listedPVZRow

	for rows.Next() {
		var row listedPVZRow
		if err := rows.StructScan(&row); err != nil {
			r.logger.Error("failed to scan PVZ row", zap.Error(err))
			return nil, databaseerrors.ErrUnexpected
		}

		rawRows = append(rawRows, &row)
	}

	return r.listedToModel(rawRows), nil
}

// GetAll возвращает все существующие ПВЗ из базы данных.
// Возвращает список доменных моделей models.PVZ или ошибку, если она была
func (r *postgresqlPVZRepository) GetAll(ctx context.Context) ([]*models.PVZ, error) {
	query := `SELECT id, registration_date, city FROM pvzs`

	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		r.logger.Error("failed to get all PVZs", zap.Error(err))
		return nil, databaseerrors.ErrUnexpected
	}
	defer rows.Close()

	pvzs := make([]*models.PVZ, 0)

	for rows.Next() {
		var row pvzRow
		if err := rows.StructScan(&row); err != nil {
			r.logger.Error("failed to scan PVZ row", zap.Error(err))
			return nil, databaseerrors.ErrUnexpected
		}

		pvzs = append(pvzs, r.toModel(row))
	}

	return pvzs, nil
}
