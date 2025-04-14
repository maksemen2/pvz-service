//go:build unit
// +build unit

package service_test

import (
	"context"
	mock_repositories "github.com/maksemen2/pvz-service/internal/domain/repositories/mocks"
	"go.uber.org/mock/gomock"
	"testing"
	"time"

	"github.com/google/uuid"
	domainerrors "github.com/maksemen2/pvz-service/internal/domain/errors"
	"github.com/maksemen2/pvz-service/internal/domain/models"
	databaseerrors "github.com/maksemen2/pvz-service/internal/repository/errors"
	"github.com/maksemen2/pvz-service/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCloseLastReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repositories.NewMockIReceptionRepo(ctrl)
	logger := zap.NewNop()
	svc := service.NewReceptionService(logger, mockRepo)

	pvzID := uuid.New()
	expectedReception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.ReceptionStatusClose,
	}

	t.Run("Successful close", func(t *testing.T) {
		mockRepo.EXPECT().CloseLast(gomock.Any(), pvzID).Return(expectedReception, nil)

		reception, err := svc.CloseLastReception(
			context.Background(),
			models.RoleEmployee.String(),
			pvzID,
		)

		assert.NoError(t, err)
		assert.Equal(t, expectedReception, reception)
	})

	t.Run("Invalid role", func(t *testing.T) {
		_, err := svc.CloseLastReception(
			context.Background(),
			models.RoleModerator.String(),
			pvzID,
		)
		assert.ErrorIs(t, err, domainerrors.ErrNotEnoughRights)
	})

	t.Run("Repository unexpected error", func(t *testing.T) {
		mockRepo.EXPECT().CloseLast(gomock.Any(), pvzID).Return(nil, databaseerrors.ErrUnexpected)

		_, err := svc.CloseLastReception(
			context.Background(),
			models.RoleEmployee.String(),
			pvzID,
		)
		assert.ErrorIs(t, err, domainerrors.ErrUnexpected)
	})

}

func TestCreateReceptionIfNoOpen(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repositories.NewMockIReceptionRepo(ctrl)
	logger := zap.NewNop()
	svc := service.NewReceptionService(logger, mockRepo)

	pvzID := uuid.New()

	t.Run("Successful create", func(t *testing.T) {
		mockRepo.EXPECT().CreateIfNoOpen(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, r *models.Reception) error {
				assert.NotEqual(t, uuid.Nil, r.ID)
				assert.False(t, r.DateTime.IsZero())
				assert.Equal(t, pvzID, r.PVZID)
				assert.Equal(t, models.ReceptionStatusInProgress, r.Status)

				return nil
			},
		)

		reception, err := svc.CreateReceptionIfNoOpen(
			context.Background(),
			models.RoleEmployee.String(),
			pvzID,
		)

		assert.NoError(t, err)
		assert.NotNil(t, reception)
		assert.True(t, time.Since(reception.DateTime) < time.Second)
	})

	t.Run("Invalid role", func(t *testing.T) {
		_, err := svc.CreateReceptionIfNoOpen(
			context.Background(),
			models.RoleModerator.String(),
			pvzID,
		)
		assert.ErrorIs(t, err, domainerrors.ErrNotEnoughRights)
	})

	t.Run("Repository unexpected error", func(t *testing.T) {
		mockRepo.EXPECT().CreateIfNoOpen(gomock.Any(), gomock.Any()).Return(databaseerrors.ErrUnexpected)

		_, err := svc.CreateReceptionIfNoOpen(
			context.Background(),
			models.RoleEmployee.String(),
			pvzID,
		)
		assert.ErrorIs(t, err, domainerrors.ErrUnexpected)
	})

	t.Run("No rows repository error", func(t *testing.T) {
		mockRepo.EXPECT().CreateIfNoOpen(gomock.Any(), gomock.Any()).Return(databaseerrors.ErrNoRows)

		_, err := svc.CreateReceptionIfNoOpen(
			context.Background(),
			models.RoleEmployee.String(),
			pvzID,
		)
		assert.ErrorIs(t, err, domainerrors.ErrPVZNotFound)
	})
}
