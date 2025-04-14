//go:build unit
// +build unit

package httphandlers_test

import (
	"bytes"
	"encoding/json"
	httphandlers "github.com/maksemen2/pvz-service/internal/delivery/http/handlers"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/maksemen2/pvz-service/internal/delivery/http/httpdto"
	domainerrors "github.com/maksemen2/pvz-service/internal/domain/errors"
	"github.com/maksemen2/pvz-service/internal/domain/models"
	"github.com/maksemen2/pvz-service/internal/pkg/auth"
	service_mocks "github.com/maksemen2/pvz-service/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestProductHandler_HandleDeleteLastProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductService := service_mocks.NewMockProductService(ctrl)
	logger := zap.NewNop()

	tests := []struct {
		name         string
		pvzID        string
		role         models.RoleType
		mockSetup    func(string)
		expectedCode int
	}{
		{
			name:  "successful delete",
			pvzID: uuid.New().String(),
			role:  models.RoleEmployee,
			mockSetup: func(pvzID string) {
				pvzUUID := uuid.MustParse(pvzID)
				mockProductService.EXPECT().
					DeleteLastProduct(gomock.Any(), gomock.Eq(string(models.RoleEmployee)), pvzUUID).
					Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid pvzID format",
			pvzID:        "invalid-uuid",
			role:         models.RoleEmployee,
			mockSetup:    func(pvzID string) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:  "insufficient permissions",
			pvzID: uuid.New().String(),
			role:  models.RoleEmployee,
			mockSetup: func(pvzID string) {
				mockProductService.EXPECT().
					DeleteLastProduct(gomock.Any(), gomock.Eq(string(models.RoleEmployee)), gomock.Any()).
					Return(domainerrors.ErrNotEnoughRights)
			},
			expectedCode: http.StatusForbidden,
		},
		{
			name:  "no open receptions",
			pvzID: uuid.New().String(),
			role:  models.RoleEmployee,
			mockSetup: func(pvzID string) {
				mockProductService.EXPECT().
					DeleteLastProduct(gomock.Any(), gomock.Eq(string(models.RoleEmployee)), gomock.Any()).
					Return(domainerrors.ErrNoOpenReceptions)
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(tt.pvzID)

			handler := httphandlers.NewProductHandler(logger, mockProductService)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/pvz/:pvzId/delete_last_product", func(c *gin.Context) {
				c.Set(auth.RoleKey, string(tt.role))
				handler.HandleDeleteLastProduct(c)
			})

			req, _ := http.NewRequest("POST", "/pvz/"+tt.pvzID+"/delete_last_product", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedCode, resp.Code)
		})
	}
}

func TestProductHandler_HandleAddProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductService := service_mocks.NewMockProductService(ctrl)
	logger := zap.NewNop()

	validPvzID := uuid.New()
	validReceptionID := uuid.New()
	validProduct := &models.Product{
		ID:          uuid.New(),
		Type:        models.ProductTypeClothes,
		ReceptionID: validReceptionID,
	}

	tests := []struct {
		name         string
		requestBody  interface{}
		role         models.RoleType
		mockSetup    func()
		expectedCode int
	}{
		{
			name: "Successful product add",
			requestBody: httpdto.PostProductsJSONRequestBody{
				Type:  "одежда",
				PvzId: validPvzID,
			},
			role: models.RoleEmployee,
			mockSetup: func() {
				mockProductService.EXPECT().
					AddProduct(gomock.Any(), gomock.Eq(string(models.RoleEmployee)), models.ProductTypeClothes.String(), validPvzID).
					Return(validProduct, nil)
			},
			expectedCode: http.StatusCreated,
		},
		{
			name:         "Invalid request body",
			requestBody:  "invalid",
			role:         models.RoleEmployee,
			mockSetup:    func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Invalid product type",
			requestBody: httpdto.PostProductsJSONRequestBody{
				Type:  "invalid_type",
				PvzId: validPvzID,
			},
			role: models.RoleEmployee,
			mockSetup: func() {
				mockProductService.EXPECT().
					AddProduct(gomock.Any(), gomock.Eq(string(models.RoleEmployee)), "invalid_type", validPvzID).
					Return(nil, domainerrors.ErrInvalidProductType)
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Not enough rights",
			requestBody: httpdto.PostProductsJSONRequestBody{
				Type:  "одежда",
				PvzId: validPvzID,
			},
			role: models.RoleEmployee,
			mockSetup: func() {
				mockProductService.EXPECT().
					AddProduct(gomock.Any(), gomock.Eq(string(models.RoleEmployee)), "одежда", validPvzID).
					Return(nil, domainerrors.ErrNotEnoughRights)
			},
			expectedCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			handler := httphandlers.NewProductHandler(logger, mockProductService)

			gin.SetMode(gin.TestMode)
			router := gin.New()

			router.POST("/products", func(c *gin.Context) {
				c.Set(auth.RoleKey, string(tt.role))
				handler.HandleAddProduct(c)
			})

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(body))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedCode, resp.Code)
		})
	}
}
