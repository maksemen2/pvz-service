package repositories

import (
	"context"
	"github.com/maksemen2/pvz-service/internal/domain/models"
)

// IPVZRepo - интерфейс для репозитория ПВЗ.
type IPVZRepo interface {
	Create(ctx context.Context, pvz *models.PVZ) error                                       // Создает запись о ПВЗ из доменной модели и возвращает ошибку.
	List(ctx context.Context, filter *models.PVZFilter) ([]*models.PVZWithReceptions, error) // Возвращает список ПВЗ по фильтрам и возвращает ПВЗ с приемками в них с товарами в них.
	GetAll(ctx context.Context) ([]*models.PVZ, error)                                       // Получает все ПВЗ из базы даных
}
