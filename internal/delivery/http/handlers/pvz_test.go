//go:build unit
// +build unit

package httphandlers_test

import (
	"bytes"
	"encoding/json"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	httphandlers "github.com/maksemen2/pvz-service/internal/delivery/http/handlers"
	"github.com/maksemen2/pvz-service/internal/delivery/http/httpdto"
	domainerrors "github.com/maksemen2/pvz-service/internal/domain/errors"
	"github.com/maksemen2/pvz-service/internal/domain/models"
	"github.com/maksemen2/pvz-service/internal/pkg/auth"
	service_mocks "github.com/maksemen2/pvz-service/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestPVZHandler_HandleCreatePVZ(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPVZService := service_mocks.NewMockPVZService(ctrl)
	logger := zap.NewNop()

	now := time.Now()
	pvzID := uuid.New()

	tests := []struct {
		name         string
		requestBody  interface{}
		role         models.RoleType
		mockSetup    func()
		expectedCode int
	}{
		{
			name: "Successful create",
			requestBody: httpdto.PostPvzJSONRequestBody{
				City:             "Москва",
				Id:               &pvzID,
				RegistrationDate: &now,
			},
			role: models.RoleModerator,
			mockSetup: func() {
				mockPVZService.EXPECT().
					CreatePVZ(gomock.Any(), models.RoleModerator.String(), "Москва", &pvzID, gomock.Any()).
					Return(&models.PVZ{ID: pvzID, City: "Москва"}, nil)
			},
			expectedCode: http.StatusCreated,
		},
		{
			name:         "Invalid request body",
			requestBody:  "invalid",
			role:         models.RoleModerator,
			mockSetup:    func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Invalid city",
			requestBody: httpdto.PostPvzJSONRequestBody{
				City:             "invalid_city",
				Id:               &pvzID,
				RegistrationDate: &now,
			},
			role: models.RoleModerator,
			mockSetup: func() {
				mockPVZService.EXPECT().
					CreatePVZ(gomock.Any(), models.RoleModerator.String(), "invalid_city", &pvzID, gomock.Any()).
					Return(nil, domainerrors.ErrInvalidCity)
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Not enough rights",
			requestBody: httpdto.PostPvzJSONRequestBody{
				City:             "Москва",
				Id:               &pvzID,
				RegistrationDate: &now,
			},
			role: models.RoleEmployee,
			mockSetup: func() {
				mockPVZService.EXPECT().
					CreatePVZ(gomock.Any(), models.RoleEmployee.String(), "Москва", &pvzID, gomock.Any()).
					Return(nil, domainerrors.ErrUserNotModerator)
			},
			expectedCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			handler := httphandlers.NewPVZHandler(logger, mockPVZService)

			gin.SetMode(gin.TestMode)
			router := gin.New()

			router.POST("/pvz", func(c *gin.Context) {
				c.Set(auth.RoleKey, string(tt.role))
				handler.HandleCreatePVZ(c)
			})

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/pvz", bytes.NewBuffer(body))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedCode, resp.Code)
		})
	}
}

func TestPVZHandler_HandleListPVZ(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPVZService := service_mocks.NewMockPVZService(ctrl)
	logger := zap.NewNop()

	now := time.Now()
	pvzID := uuid.New()

	page := 1
	limit := 10

	tests := []struct {
		name         string
		queryParams  map[string]string
		role         models.RoleType
		mockSetup    func()
		expectedCode int
	}{
		{
			name: "Successful list",
			queryParams: map[string]string{
				"startDate": now.Format(time.RFC3339),
				"endDate":   now.Add(24 * time.Hour).Format(time.RFC3339),
				"page":      "1",
				"limit":     "10",
			},
			role: models.RoleModerator,
			mockSetup: func() {
				mockPVZService.EXPECT().
					ListPVZs(gomock.Any(), models.RoleModerator.String(), gomock.Any(), gomock.Any(), &page, &limit).
					Return([]*models.PVZWithReceptions{
						{PVZ: &models.PVZ{ID: pvzID, City: "Москва"}},
					}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Invalid date",
			queryParams: map[string]string{
				"startDate": "invalid_date",
			},
			role:         models.RoleModerator,
			mockSetup:    func() {},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			handler := httphandlers.NewPVZHandler(logger, mockPVZService)

			gin.SetMode(gin.TestMode)
			router := gin.New()

			router.GET("/pvz", func(c *gin.Context) {
				c.Set(auth.RoleKey, string(tt.role))
				handler.HandleListPVZ(c)
			})

			req, _ := http.NewRequest("GET", "/pvz", nil)

			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}

			req.URL.RawQuery = q.Encode()

			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedCode, resp.Code)
		})
	}
}
