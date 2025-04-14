//go:build unit
// +build unit

package service_test

import (
	"context"
	"github.com/google/uuid"
	domainerrors "github.com/maksemen2/pvz-service/internal/domain/errors"
	"github.com/maksemen2/pvz-service/internal/domain/models"
	mock_repositories "github.com/maksemen2/pvz-service/internal/domain/repositories/mocks"
	databaseerrors "github.com/maksemen2/pvz-service/internal/repository/errors"
	"github.com/maksemen2/pvz-service/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"testing"
)

func TestAddProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repositories.NewMockIProductRepo(ctrl)
	logger := zap.NewNop()
	svc := service.NewProductService(logger, mockRepo)

	pvzID := uuid.New()
	productType := "электроника"

	t.Run("Successful add", func(t *testing.T) {
		mockRepo.
			EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, addProduct *models.AddProduct) (*models.Product, error) {
				assert.Equal(t, models.ProductTypeElectronics, addProduct.Type)
				assert.Equal(t, pvzID, addProduct.PVZID)
				assert.NotEqual(t, uuid.Nil, addProduct.ID)
				return &models.Product{
					ID:          addProduct.ID,
					DateTime:    addProduct.DateTime,
					Type:        addProduct.Type,
					ReceptionID: pvzID,
				}, nil
			})

		product, err := svc.AddProduct(
			context.Background(),
			models.RoleEmployee.String(),
			productType,
			pvzID,
		)

		assert.NoError(t, err)
		assert.Equal(t, models.ProductTypeElectronics, product.Type)
		assert.Equal(t, pvzID, product.ReceptionID)
		assert.NotEqual(t, uuid.Nil, product.ID)
	})

	// Только сотрудник может добавлять товары
	t.Run("Invalid role", func(t *testing.T) {
		_, err := svc.AddProduct(
			context.Background(),
			models.RoleModerator.String(),
			productType,
			pvzID,
		)
		assert.ErrorIs(t, err, domainerrors.ErrNotEnoughRights)
	})

	t.Run("Invalid product type", func(t *testing.T) {
		_, err := svc.AddProduct(
			context.Background(),
			models.RoleEmployee.String(),
			"invalid_type",
			pvzID,
		)
		assert.ErrorContains(t, err, domainerrors.ErrInvalidProductType.Error())
	})

	t.Run("Repository error", func(t *testing.T) {
		mockRepo.
			EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil, databaseerrors.ErrUnexpected)

		_, err := svc.AddProduct(
			context.Background(),
			models.RoleEmployee.String(),
			productType,
			pvzID,
		)
		assert.ErrorIs(t, err, domainerrors.ErrUnexpected)
	})
}

func TestDeleteLastProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repositories.NewMockIProductRepo(ctrl)
	logger := zap.NewNop()
	svc := service.NewProductService(logger, mockRepo)

	pvzID := uuid.New()

	t.Run("Successful delete", func(t *testing.T) {
		mockRepo.EXPECT().DeleteLast(gomock.Any(), pvzID).Return(nil)

		err := svc.DeleteLastProduct(
			context.Background(),
			models.RoleEmployee.String(),
			pvzID,
		)
		assert.NoError(t, err)
	})

	t.Run("Invalid role", func(t *testing.T) {
		err := svc.DeleteLastProduct(
			context.Background(),
			models.RoleModerator.String(),
			pvzID,
		)
		assert.ErrorIs(t, err, domainerrors.ErrNotEnoughRights)
	})

	t.Run("Repository error", func(t *testing.T) {
		mockRepo.EXPECT().DeleteLast(gomock.Any(), pvzID).Return(databaseerrors.ErrUnexpected)

		err := svc.DeleteLastProduct(
			context.Background(),
			models.RoleEmployee.String(),
			pvzID,
		)
		assert.ErrorIs(t, err, domainerrors.ErrUnexpected)
	})
}
