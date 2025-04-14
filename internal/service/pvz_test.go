//go:build unit
// +build unit

package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	domainerrors "github.com/maksemen2/pvz-service/internal/domain/errors"
	"github.com/maksemen2/pvz-service/internal/domain/models"
	"github.com/maksemen2/pvz-service/internal/domain/repositories/mocks"
	databaseerrors "github.com/maksemen2/pvz-service/internal/repository/errors"
	"github.com/maksemen2/pvz-service/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestCreatePVZ(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repositories.NewMockIPVZRepo(ctrl)
	logger := zap.NewNop()
	svc := service.NewPVZService(logger, mockRepo)

	now := time.Now()
	testUUID := uuid.New()
	validCity := models.CityTypeMoscow

	// Кейс когда айди и дата создания уже даны
	t.Run("Success with provided ID and date", func(t *testing.T) {
		expectedPVZ := &models.PVZ{
			ID:               testUUID,
			City:             validCity,
			RegistrationDate: now,
		}

		mockRepo.EXPECT().Create(gomock.Any(), expectedPVZ).Return(nil)

		pvz, err := svc.CreatePVZ(
			context.Background(),
			models.RoleModerator.String(),
			string(validCity),
			&testUUID,
			&now,
		)

		assert.NoError(t, err)
		assert.Equal(t, expectedPVZ, pvz)
	})

	// Кейс когда айди и дата не переданы
	t.Run("Success with generated ID and date", func(t *testing.T) {
		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, pvz *models.PVZ) error {
				assert.NotEqual(t, uuid.Nil, pvz.ID)
				assert.False(t, pvz.RegistrationDate.IsZero())

				return nil
			},
		)

		pvz, err := svc.CreatePVZ(
			context.Background(),
			models.RoleModerator.String(),
			validCity.String(),
			nil,
			nil,
		)

		assert.NoError(t, err)
		assert.NotNil(t, pvz)
	})

	t.Run("Invalid city", func(t *testing.T) {
		_, err := svc.CreatePVZ(
			context.Background(),
			models.RoleModerator.String(),
			"invalid_city",
			nil,
			nil,
		)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidCity)
	})

	// employee не имеет права создавать пвз
	t.Run("Invalid role", func(t *testing.T) {
		_, err := svc.CreatePVZ(
			context.Background(),
			models.RoleEmployee.String(),
			validCity.String(),
			nil,
			nil,
		)
		assert.ErrorIs(t, err, domainerrors.ErrUserNotModerator)
	})

	// Кейс когда дали айди уже существующего ПВЗ
	t.Run("Repository unique violation error", func(t *testing.T) {
		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(databaseerrors.ErrUniqueViolation)

		_, err := svc.CreatePVZ(
			context.Background(),
			models.RoleModerator.String(),
			validCity.String(),
			&testUUID,
			&now,
		)
		assert.ErrorIs(t, err, domainerrors.ErrPVZAlreadyExists)
	})

	t.Run("Repository unexpected error", func(t *testing.T) {
		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(databaseerrors.ErrUnexpected)

		_, err := svc.CreatePVZ(
			context.Background(),
			models.RoleModerator.String(),
			validCity.String(),
			&testUUID,
			&now,
		)
		assert.ErrorIs(t, err, domainerrors.ErrUnexpected)
	})
}

func TestListPVZs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repositories.NewMockIPVZRepo(ctrl)
	logger := zap.NewNop()
	svc := service.NewPVZService(logger, mockRepo)

	now := time.Now()
	testPVZs := []*models.PVZWithReceptions{
		{
			PVZ: &models.PVZ{
				ID:               uuid.New(),
				City:             models.CityType("москва"),
				RegistrationDate: now,
			},
		},
	}

	t.Run("Success with filters", func(t *testing.T) {
		start := now.Add(-24 * time.Hour)
		end := now
		page := 2
		limit := 20

		expectedFilter := &models.PVZFilter{
			StartDate: &start,
			EndDate:   &end,
			Page:      page,
			PageSize:  limit,
		}

		mockRepo.EXPECT().List(gomock.Any(), expectedFilter).Return(testPVZs, nil)

		result, err := svc.ListPVZs(
			context.Background(),
			models.RoleModerator.String(),
			&start,
			&end,
			&page,
			&limit,
		)

		assert.NoError(t, err)
		assert.Equal(t, testPVZs, result)
	})

	t.Run("Invalid role", func(t *testing.T) {
		_, err := svc.ListPVZs(
			context.Background(),
			"invalid_role",
			nil,
			nil,
			nil,
			nil,
		)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidRole)
	})

	t.Run("Invalid filter (negative page)", func(t *testing.T) {
		page := -1
		_, err := svc.ListPVZs(
			context.Background(),
			models.RoleEmployee.String(),
			nil,
			nil,
			&page,
			nil,
		)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidPage)
	})

	t.Run("Default pagination values", func(t *testing.T) {
		expectedFilter := &models.PVZFilter{
			Page:     1,
			PageSize: 10,
		}

		mockRepo.EXPECT().List(gomock.Any(), expectedFilter).Return(testPVZs, nil)

		result, err := svc.ListPVZs(
			context.Background(),
			models.RoleEmployee.String(),
			nil,
			nil,
			nil,
			nil,
		)

		assert.NoError(t, err)
		assert.Equal(t, testPVZs, result)
	})

	t.Run("Repository error", func(t *testing.T) {
		mockRepo.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, databaseerrors.ErrUnexpected)

		_, err := svc.ListPVZs(
			context.Background(),
			models.RoleEmployee.String(),
			nil,
			nil,
			nil,
			nil,
		)
		assert.ErrorIs(t, err, domainerrors.ErrUnexpected)
	})
}

func TestGetAllPVZs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repositories.NewMockIPVZRepo(ctrl)
	logger := zap.NewNop()
	svc := service.NewPVZService(logger, mockRepo)

	testPVZs := []*models.PVZ{
		{
			ID:               uuid.New(),
			City:             models.CityTypeMoscow,
			RegistrationDate: time.Now(),
		},
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().GetAll(gomock.Any()).Return(testPVZs, nil)

		result, err := svc.GetAllPVZs(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, testPVZs, result)
	})

	t.Run("Repository error", func(t *testing.T) {
		mockRepo.EXPECT().GetAll(gomock.Any()).Return(nil, databaseerrors.ErrUnexpected)

		_, err := svc.GetAllPVZs(context.Background())
		assert.ErrorIs(t, err, domainerrors.ErrUnexpected)
	})
}
