package repositories

import (
	"context"
	"github.com/google/uuid"
	"github.com/maksemen2/pvz-service/internal/domain/models"
)

// IProductRepo - интерфейс для репозитория товаров.
type IProductRepo interface {
	Create(ctx context.Context, product *models.AddProduct) (*models.Product, error) // Создает запись о товаре из доменной модели и возвращает ошибку.
	DeleteLast(ctx context.Context, pvzID uuid.UUID) error                           // Удаляет последнюю запись о товаре из последней открытой приёмки указанного PVZ и возвращает ошибку.
}
