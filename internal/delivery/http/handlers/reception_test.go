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

func TestReceptionHandler_HandleCloseLastReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReceptionService := service_mocks.NewMockReceptionService(ctrl)
	logger := zap.NewNop()

	tests := []struct {
		name         string
		pvzID        string
		role         models.RoleType
		mockSetup    func(string)
		expectedCode int
	}{
		{
			name:  "Successful close",
			pvzID: uuid.New().String(),
			role:  models.RoleModerator,
			mockSetup: func(pvzID string) {
				pvzUUID := uuid.MustParse(pvzID)
				mockReceptionService.EXPECT().
					CloseLastReception(gomock.Any(), models.RoleModerator.String(), pvzUUID).
					Return(&models.Reception{ID: uuid.New()}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid pvz id format",
			pvzID:        "invalid-uuid",
			role:         models.RoleModerator,
			mockSetup:    func(string) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:  "No open receptions",
			pvzID: uuid.New().String(),
			role:  models.RoleModerator,
			mockSetup: func(pvzID string) {
				pvzUUID := uuid.MustParse(pvzID)
				mockReceptionService.EXPECT().
					CloseLastReception(gomock.Any(), models.RoleModerator.String(), pvzUUID).
					Return(nil, domainerrors.ErrNoOpenReceptions)
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:  "Not enough rights",
			pvzID: uuid.New().String(),
			role:  models.RoleEmployee,
			mockSetup: func(pvzID string) {
				pvzUUID := uuid.MustParse(pvzID)
				mockReceptionService.EXPECT().
					CloseLastReception(gomock.Any(), models.RoleEmployee.String(), pvzUUID).
					Return(nil, domainerrors.ErrNotEnoughRights)
			},
			expectedCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(tt.pvzID)

			handler := httphandlers.NewReceptionHandler(logger, mockReceptionService)

			gin.SetMode(gin.TestMode)
			router := gin.New()

			router.POST("/pvz/:pvzId/close_last_reception", func(c *gin.Context) {
				c.Set(auth.RoleKey, tt.role.String())
				handler.HandleCloseLastReception(c)
			})

			req, _ := http.NewRequest("POST", "/pvz/"+tt.pvzID+"/close_last_reception", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedCode, resp.Code)
		})
	}
}

func TestReceptionHandler_HandleCreateReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReceptionService := service_mocks.NewMockReceptionService(ctrl)
	logger := zap.NewNop()

	validPvzID := uuid.New()
	validReception := &models.Reception{
		ID:    uuid.New(),
		PVZID: validPvzID,
	}

	tests := []struct {
		name         string
		requestBody  interface{}
		role         models.RoleType
		mockSetup    func()
		expectedCode int
	}{
		{
			name: "Successful creation",
			requestBody: httpdto.PostReceptionsJSONRequestBody{
				PvzId: validPvzID,
			},
			role: models.RoleEmployee,
			mockSetup: func() {
				mockReceptionService.EXPECT().
					CreateReceptionIfNoOpen(gomock.Any(), models.RoleEmployee.String(), validPvzID).
					Return(validReception, nil)
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
			name: "Open reception exist",
			requestBody: httpdto.PostReceptionsJSONRequestBody{
				PvzId: validPvzID,
			},
			role: models.RoleEmployee,
			mockSetup: func() {
				mockReceptionService.EXPECT().
					CreateReceptionIfNoOpen(gomock.Any(), models.RoleEmployee.String(), validPvzID).
					Return(nil, domainerrors.ErrOpenReceptionExists)
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Not enough rights",
			requestBody: httpdto.PostReceptionsJSONRequestBody{
				PvzId: validPvzID,
			},
			role: models.RoleEmployee,
			mockSetup: func() {
				mockReceptionService.EXPECT().
					CreateReceptionIfNoOpen(gomock.Any(), models.RoleEmployee.String(), validPvzID).
					Return(nil, domainerrors.ErrNotEnoughRights)
			},
			expectedCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			handler := httphandlers.NewReceptionHandler(logger, mockReceptionService)

			gin.SetMode(gin.TestMode)
			router := gin.New()

			router.POST("/receptions", func(c *gin.Context) {
				c.Set(auth.RoleKey, tt.role.String())
				handler.HandleCreateReception(c)
			})

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/receptions", bytes.NewBuffer(body))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedCode, resp.Code)
		})
	}
}
