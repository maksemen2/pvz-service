package repositories

import (
	"context"

	"github.com/maksemen2/pvz-service/internal/domain/models"
)

// IUserRepo - интерфейс для репозитория пользователей.
type IUserRepo interface {
	Create(ctx context.Context, user *models.User) error                // Создает запись о пользователе из доменной модели и возвращает ошибку.
	GetByEmail(ctx context.Context, email string) (*models.User, error) // Находит пользователя по email и возвращает его доменную модель и ошибку.
}
