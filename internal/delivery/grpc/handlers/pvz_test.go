//go:build unit
// +build unit

package grpchandlers_test

import (
	"context"
	"github.com/google/uuid"
	grpchandlers "github.com/maksemen2/pvz-service/internal/delivery/grpc/handlers"
	domainerrors "github.com/maksemen2/pvz-service/internal/domain/errors"
	"github.com/maksemen2/pvz-service/internal/domain/models"
	service_mocks "github.com/maksemen2/pvz-service/internal/service/mocks"
	"go.uber.org/mock/gomock"
	"testing"
	"time"

	"github.com/maksemen2/pvz-service/internal/delivery/grpc/pvz_v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestPVZServer_GetPVZList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := service_mocks.NewMockPVZService(ctrl)
	handler := grpchandlers.NewPVZServer(mockService)

	ctx := context.Background()

	t.Run("Successful list", func(t *testing.T) {
		expectedPVZs := []*models.PVZ{
			{ID: uuid.New(), RegistrationDate: time.Now(), City: models.CityTypeKazan},
			{ID: uuid.New(), RegistrationDate: time.Now(), City: models.CityTypeMoscow},
		}

		mockService.EXPECT().
			GetAllPVZs(ctx).
			Return(expectedPVZs, nil).
			Times(1)

		resp, err := handler.GetPVZList(ctx, &pvz_v1.GetPVZListRequest{})

		assert.NoError(t, err)
		assert.Len(t, resp.Pvzs, 2)
		assert.Equal(t, models.CityTypeKazan.String(), resp.Pvzs[0].City)
		assert.Equal(t, models.CityTypeMoscow.String(), resp.Pvzs[1].City)
	})

	t.Run("Service error", func(t *testing.T) {

		mockService.EXPECT().
			GetAllPVZs(ctx).
			Return(nil, domainerrors.ErrUnexpected).
			Times(1)

		resp, err := handler.GetPVZList(ctx, &pvz_v1.GetPVZListRequest{})

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}
