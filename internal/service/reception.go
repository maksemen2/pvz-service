package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	domainerrors "github.com/maksemen2/pvz-service/internal/domain/errors"
	"github.com/maksemen2/pvz-service/internal/domain/models"
	"github.com/maksemen2/pvz-service/internal/domain/repositories"
	"github.com/maksemen2/pvz-service/internal/pkg/metrics"
	databaseerrors "github.com/maksemen2/pvz-service/internal/repository/errors"
	"go.uber.org/zap"
	"time"
)

// ReceptionService - интерфейс для работы с приемами ПВЗ.
type ReceptionService interface {
	CloseLastReception(ctx context.Context, userRole string, pvzID uuid.UUID) (*models.Reception, error)
	CreateReceptionIfNoOpen(ctx context.Context, userRole string, pvzID uuid.UUID) (*models.Reception, error)
}

// receptionServiceImpl реализует интерфейс ReceptionService.
type receptionServiceImpl struct {
	logger *zap.Logger
	repo   repositories.IReceptionRepo
}

// NewReceptionService - конструктор для создания нового экземпляра ReceptionService.
// Принимает логгер и репозиторий приемок.
func NewReceptionService(logger *zap.Logger, repo repositories.IReceptionRepo) ReceptionService {
	return &receptionServiceImpl{
		logger: logger,
		repo:   repo,
	}
}

// CloseLastReception закрывает последнюю приемку в ПВЗ.
// Принимает роль пользователя и айди ПВЗ.
// Проводит валидацию роли пользователя (только models.RoleEmployee может закрывать приемки).
// Возвращает закрытую приемку с обновленными данными о ней и ошибку, если она возникла.
func (s *receptionServiceImpl) CloseLastReception(ctx context.Context, userRole string, pvzID uuid.UUID) (*models.Reception, error) {
	userRoleType := models.RoleType(userRole)
	if userRoleType != models.RoleEmployee {
		return nil, domainerrors.ErrNotEnoughRights
	}

	reception, err := s.repo.CloseLast(ctx, pvzID)

	if err != nil {
		if errors.Is(err, databaseerrors.ErrUnexpected) {
			return nil, domainerrors.ErrUnexpected
		}

		return nil, err // Репозиторий может возвращать и доменные ошибки
	}

	return reception, nil
}

// CreateReceptionIfNoOpen создает новую приемку, если в ПВЗ нет открытой приемки.
// Принимает роль пользователя и айди ПВЗ.
// Проводит валидацию роли пользователя (только models.RoleEmployee может создавать приемки).
// Возвращает созданную приемку и ошибку, если она возникла.
func (s *receptionServiceImpl) CreateReceptionIfNoOpen(ctx context.Context, userRole string, pvzID uuid.UUID) (*models.Reception, error) {
	userRoleType := models.RoleType(userRole)
	if userRoleType != models.RoleEmployee {
		return nil, domainerrors.ErrNotEnoughRights
	}

	reception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.ReceptionStatusInProgress,
	}

	err := s.repo.CreateIfNoOpen(ctx, reception)

	if err != nil {
		switch {
		case errors.Is(err, databaseerrors.ErrUnexpected):
			return nil, domainerrors.ErrUnexpected
		case errors.Is(err, databaseerrors.ErrNoRows):
			return nil, domainerrors.ErrPVZNotFound
		}

		return nil, err
	}

	metrics.ReceptionsCreated.Inc()

	return reception, nil
}
