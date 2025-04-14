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
	"github.com/maksemen2/pvz-service/internal/domain/errors"
	"github.com/maksemen2/pvz-service/internal/domain/models"
	service_mocks "github.com/maksemen2/pvz-service/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAuthHandler_HandleRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := service_mocks.NewMockAuthService(ctrl)
	logger := zap.NewNop()

	tests := []struct {
		name         string
		requestBody  interface{}
		mockSetup    func()
		expectedCode int
		expectedUser httpdto.User
	}{
		{
			name: "Successful register",
			requestBody: httpdto.PostRegisterJSONBody{
				Email:    "test@example.com",
				Password: "password",
				Role:     "employee",
			},
			mockSetup: func() {
				user := &models.User{
					ID:    uuid.New(),
					Email: "test@example.com",
					Role:  models.RoleEmployee,
				}
				mockAuthService.EXPECT().RegisterUser(gomock.Any(), "test@example.com", "password", "employee").Return(user, nil)
			},
			expectedCode: http.StatusCreated,
			expectedUser: httpdto.User{
				Email: "test@example.com",
				Role:  "employee",
			},
		},
		{
			name:         "Invalid request body",
			requestBody:  "invalid",
			mockSetup:    func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "User exist",
			requestBody: httpdto.PostRegisterJSONBody{
				Email:    "exists@example.com",
				Password: "password",
				Role:     "employee",
			},
			mockSetup: func() {
				mockAuthService.EXPECT().
					RegisterUser(gomock.Any(), "exists@example.com", "password", "employee").
					Return(nil, domainerrors.ErrUserExists)
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			handler := httphandlers.NewAuthHandler(logger, mockAuthService)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/register", handler.HandleRegister)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)
			assert.Equal(t, tt.expectedCode, resp.Code)
		})
	}
}

func TestAuthHandler_HandleLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := service_mocks.NewMockAuthService(ctrl)
	logger := zap.NewNop()

	tests := []struct {
		name         string
		requestBody  interface{}
		mockSetup    func()
		expectedCode int
	}{
		{
			name: "Successful login",
			requestBody: httpdto.PostLoginJSONBody{
				Email:    "user@example.com",
				Password: "password",
			},
			mockSetup: func() {
				mockAuthService.EXPECT().
					AuthenticateUser(gomock.Any(), "user@example.com", "password").
					Return(models.Token("valid_token"), nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Invalid credientals",
			requestBody: httpdto.PostLoginJSONBody{
				Email:    "user@example.com",
				Password: "wrong",
			},
			mockSetup: func() {
				mockAuthService.EXPECT().
					AuthenticateUser(gomock.Any(), "user@example.com", "wrong").
					Return(models.Token(""), domainerrors.ErrInvalidCredentials)
			},
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			handler := httphandlers.NewAuthHandler(logger, mockAuthService)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/login", handler.HandleLogin)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)
			assert.Equal(t, tt.expectedCode, resp.Code)
		})
	}
}

func TestAuthHandler_HandleDummyLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := service_mocks.NewMockAuthService(ctrl)
	logger := zap.NewNop()

	tests := []struct {
		name         string
		requestBody  interface{}
		mockSetup    func()
		expectedCode int
	}{
		{
			name: "Successful dummy login",
			requestBody: httpdto.PostDummyLoginJSONBody{
				Role: "admin",
			},
			mockSetup: func() {
				mockAuthService.EXPECT().
					DummyLogin(gomock.Any(), "admin").
					Return(models.Token("dummy_token"), nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Invalid role",
			requestBody: httpdto.PostDummyLoginJSONBody{
				Role: "invalid_role",
			},
			mockSetup: func() {
				mockAuthService.EXPECT().
					DummyLogin(gomock.Any(), "invalid_role").
					Return(models.Token(""), domainerrors.ErrInvalidRole)
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			handler := httphandlers.NewAuthHandler(logger, mockAuthService)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/dummyLogin", handler.HandleDummyLogin)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/dummyLogin", bytes.NewBuffer(body))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)
			assert.Equal(t, tt.expectedCode, resp.Code)
		})
	}
}
