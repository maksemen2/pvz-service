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

// PVZService - интерфейс для бизнес-логики работы с ПВЗ (пунктами выдачи заказов).
type PVZService interface {
	CreatePVZ(ctx context.Context, userRole, city string, pvzID *uuid.UUID, registerDate *time.Time) (*models.PVZ, error)                      // Создает ПВЗ с указанием города, опциональных айди и даты регистрации.
	ListPVZs(ctx context.Context, userRole string, startDate, endDate *time.Time, pageNumber, limit *int) ([]*models.PVZWithReceptions, error) // Возвращает список ПВЗ с приемками внутри них и товарами внутри приёмок.
	GetAllPVZs(ctx context.Context) ([]*models.PVZ, error)                                                                                     // Возвращает все ПВЗ из базы данных.
}

// pvzServiceImpl реализует интерфейс PVZService
type pvzServiceImpl struct {
	logger  *zap.Logger
	pvzRepo repositories.IPVZRepo
}

// NewPVZService - конструктор для создания нового экземпляра PVZService.
// Принимает логгер и репозиторий ПВЗ.
func NewPVZService(logger *zap.Logger, pvzRepo repositories.IPVZRepo) PVZService {
	return &pvzServiceImpl{
		logger:  logger,
		pvzRepo: pvzRepo,
	}
}

// CreatePVZ создает новый ПВЗ. Принимает роль пользователя, город и опциональные pvzID и registerData.
// Производит валидацию роли пользователя (только models.RoleModerator может создавать ПВЗ)
// Производит валидацию города (см. models.CityType)
// Если pvzID или registerDate не указаны - создает новые значения (uuid.New() и time.Now()).
// Возвращает доменную модель ПВЗ или ошибку, если не удалось создать ПВЗ.
func (p *pvzServiceImpl) CreatePVZ(ctx context.Context, userRole, city string, pvzID *uuid.UUID, registerDate *time.Time) (*models.PVZ, error) {
	roleType := models.RoleType(userRole)
	if roleType == models.RoleEmployee {
		p.logger.Debug("User is not moderator", zap.String("userRole", userRole))
		return nil, domainerrors.ErrUserNotModerator
	}

	cityType := models.CityType(city)
	if !cityType.Valid() {
		p.logger.Debug("Invalid city type", zap.String("city", city))
		return nil, domainerrors.ErrInvalidCity
	}

	pvz := &models.PVZ{
		City: models.CityType(city),
	}

	if pvzID == nil {
		p.logger.Debug("PVZ ID is not provided")

		pvz.ID = uuid.New() // ???
	} else {
		pvz.ID = *pvzID
	}

	if registerDate == nil {
		p.logger.Debug("RegisterDate is not provided")

		pvz.RegistrationDate = time.Now() // ???
	} else {
		pvz.RegistrationDate = *registerDate
	}

	err := p.pvzRepo.Create(ctx, pvz)

	if err != nil {
		switch {
		case errors.Is(err, databaseerrors.ErrUniqueViolation):
			// TODO: непонятно, что делать в этой ситуации. На вход, согласно openapi схеме,
			// МОЖЕТ подаваться UUID ПВЗ, но не задокументировано поведение в случае,
			// если он уже существует. Пока просто возвращается ошибка.
			p.logger.Debug("PVZ already exists", zap.String("ID", pvz.ID.String()), zap.Error(err))
			return nil, domainerrors.ErrPVZAlreadyExists
		case errors.Is(err, databaseerrors.ErrUnexpected):
			return nil, domainerrors.ErrUnexpected
		}
	}

	metrics.PVZCreated.Inc()

	return pvz, nil
}

// ListPVZs возвращает список ПВЗ с приемками внутри них с товарами внутри приёмок с фильтрацией по дате ПРИЁМКИ товаров.
// Производит валидацию роли пользователя (только models.RoleEmployee и models.RoleModerator могут просматривать ПВЗ).
// Принимает так же номер страницы и размер страницы.
// Производит валидацию фильтра (см. models.PVZFilter). Возвращает ошибку в случае ошибки валидации или базы данных.
func (p *pvzServiceImpl) ListPVZs(ctx context.Context, userRole string, startDate, endDate *time.Time, pageNumber, limit *int) ([]*models.PVZWithReceptions, error) {
	roleType := models.RoleType(userRole)
	// Пока есть только две роли и, в принципе, смысла проверять нет, но это сделано на случай появления новых ролей
	if roleType != models.RoleEmployee && roleType != models.RoleModerator {
		p.logger.Debug("User is not moderator or employee", zap.String("userRole", userRole))
		return nil, fmt.Errorf("%w: %v", domainerrors.ErrInvalidRole, roleType)
	}

	filter := models.PVZFilter{StartDate: startDate, EndDate: endDate}

	if pageNumber == nil {
		filter.Page = 1 // Дефолтное значение, указанное в oapi схеме
	} else {
		filter.Page = *pageNumber
	}

	if limit == nil {
		filter.PageSize = 10 // Так же дефолтное значение, указанное в oapi схеме
	} else {
		filter.PageSize = *limit
	}

	if err := filter.Valid(); err != nil {
		// Валидация возвращает доменную ошибку, поэтому её можно сразу вернуть
		p.logger.Debug("Invalid filter", zap.Error(err))
		return nil, err
	}

	result, err := p.pvzRepo.List(ctx, &filter)

	if err != nil {
		switch {
		case errors.Is(err, databaseerrors.ErrUnexpected):
			return nil, domainerrors.ErrUnexpected
		}
	}

	return result, nil
}

// GetAllPVZs возвращает все когда-либо созданные ПВЗ.
func (p *pvzServiceImpl) GetAllPVZs(ctx context.Context) ([]*models.PVZ, error) {
	pvzs, err := p.pvzRepo.GetAll(ctx)
	if err != nil {
		switch {
		case errors.Is(err, databaseerrors.ErrUnexpected):
			return nil, domainerrors.ErrUnexpected
		}
	}

	return pvzs, nil
}
