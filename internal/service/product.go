package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	domainerrors "github.com/maksemen2/pvz-service/internal/domain/errors"
	"github.com/maksemen2/pvz-service/internal/domain/models"
	"github.com/maksemen2/pvz-service/internal/domain/repositories"
	"github.com/maksemen2/pvz-service/internal/pkg/metrics"
	databaseerrors "github.com/maksemen2/pvz-service/internal/repository/errors"
	"go.uber.org/zap"
	"time"
)

// ProductService - интерфейс для бизнес-логики работы с товарами.
type ProductService interface {
	AddProduct(ctx context.Context, userRole string, productType string, pvzID uuid.UUID) (*models.Product, error) // Добавляет товар в открытую приёмку в указанном ПВЗ
	DeleteLastProduct(ctx context.Context, userRole string, pvzID uuid.UUID) error                                 // Удаляет последний продукт из открытой приёмки в указанном ПВЗ
}

// productServiceImpl реализует интерфейс ProductService
type productServiceImpl struct {
	logger *zap.Logger
	repo   repositories.IProductRepo
}

// NewProductService - конструктор для создания нового экземпляра ProductService
// Принимает логгер и репозиторий товаров.
func NewProductService(logger *zap.Logger, repo repositories.IProductRepo) ProductService {
	return &productServiceImpl{
		logger: logger,
		repo:   repo,
	}
}

// AddProduct добавляет товар в открытую приёмку в указанном ПВЗ.
// Принимает роль пользователя, тип продукта и айди ПВЗ.
// Проводит валидацию роли пользователя (только models.RoleEmployee может добавлять товары)
// Проводит валидацию входного товара (см. models.ProductType)
// Возвращает доменную модель созданного товара или ошибку.
func (s *productServiceImpl) AddProduct(ctx context.Context, userRole string, productType string, pvzID uuid.UUID) (*models.Product, error) {
	roleType := models.RoleType(userRole)

	if roleType != models.RoleEmployee {
		return nil, domainerrors.ErrNotEnoughRights
	}

	productTypeValue := models.ProductType(productType)

	if !productTypeValue.Valid() {
		return nil, fmt.Errorf("%w: %s", domainerrors.ErrInvalidProductType, productType)
	}

	addProduct := &models.AddProduct{
		ID:       uuid.New(),
		DateTime: time.Now(),
		Type:     productTypeValue,
		PVZID:    pvzID,
	}

	product, err := s.repo.Create(ctx, addProduct)

	if err != nil {
		if errors.Is(err, databaseerrors.ErrUnexpected) {
			return nil, domainerrors.ErrUnexpected
		}

		return nil, err
	}

	metrics.ProductsAdded.Inc()

	return product, nil
}

// DeleteLastProduct удаляет последний товар из открытой приёмки в указанном ПВЗ.
// Проводит валидацию роли пользователя (только models.RoleEmployee может удалять товары)
// Возвращает ошибку, если не удалось удалить товар.
func (s *productServiceImpl) DeleteLastProduct(ctx context.Context, userRole string, pvzID uuid.UUID) error {
	roleType := models.RoleType(userRole)

	if roleType != models.RoleEmployee {
		return domainerrors.ErrNotEnoughRights
	}

	err := s.repo.DeleteLast(ctx, pvzID)

	if err != nil {
		if errors.Is(err, databaseerrors.ErrUnexpected) {
			return domainerrors.ErrUnexpected
		}

		return err
	}

	return nil
}
