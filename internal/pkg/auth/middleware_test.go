//go:build unit
// +build unit

package auth_test

import (
	mock_auth "github.com/maksemen2/pvz-service/internal/pkg/auth/mocks"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/maksemen2/pvz-service/internal/pkg/auth"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTM := mock_auth.NewMockTokenManager(ctrl)
	logger := zap.NewNop()

	router := gin.New()
	router.Use(auth.NewGinMiddleware(logger, mockTM))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	t.Run("No Authorization header", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid token format", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "InvalidToken")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Valid token format but invalid token", func(t *testing.T) {
		testToken := "bad_token"
		mockTM.EXPECT().
			Parse(testToken).
			Return(nil, auth.ErrMalformedToken)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+testToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Valid token", func(t *testing.T) {
		testUserID := uuid.New()
		testRole := "moderator"
		testToken := "good_token"

		mockClaims := mock_auth.NewMockClaims(ctrl)
		mockClaims.EXPECT().GetUserID().Return(testUserID)
		mockClaims.EXPECT().GetRole().Return(testRole)

		mockTM.EXPECT().
			Parse(testToken).
			Return(mockClaims, nil)

		var gotUserID interface{}

		var gotRole interface{}

		testRouter := gin.New()
		testRouter.Use(auth.NewGinMiddleware(logger, mockTM))
		testRouter.GET("/test", func(c *gin.Context) {
			gotUserID, _ = auth.GetUserIDFromContext(c)
			gotRole, _ = auth.GetRoleFromContext(c)
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+testToken)

		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, testUserID, gotUserID)
		assert.Equal(t, testRole, gotRole)
	})

	t.Run("Expired token", func(t *testing.T) {
		testToken := "expired_token"
		mockTM.EXPECT().
			Parse(testToken).
			Return(nil, auth.ErrTokenExpired)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+testToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
