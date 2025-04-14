package repositories

import (
	"context"
	"github.com/google/uuid"
	"github.com/maksemen2/pvz-service/internal/domain/models"
)

// IReceptionRepo - интерфейс для репозитория приемок.
type IReceptionRepo interface {
	CreateIfNoOpen(ctx context.Context, reception *models.Reception) error     // Создает запись о приемке из доменной модели и возвращает ошибку.
	CloseLast(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) // Закрывает последнюю открытую приемку для указанного PVZ и возвращает ошибку.
}
