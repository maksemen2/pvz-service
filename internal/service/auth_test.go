//go:build unit
// +build unit

package service_test

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/google/uuid"
	domainerrors "github.com/maksemen2/pvz-service/internal/domain/errors"
	"github.com/maksemen2/pvz-service/internal/domain/models"
	mock_repositories "github.com/maksemen2/pvz-service/internal/domain/repositories/mocks"
	"github.com/maksemen2/pvz-service/internal/pkg/auth"
	mock_auth "github.com/maksemen2/pvz-service/internal/pkg/auth/mocks"
	databaseerrors "github.com/maksemen2/pvz-service/internal/repository/errors"
	"github.com/maksemen2/pvz-service/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAuthService_RegisterUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repositories.NewMockIUserRepo(ctrl)
	mockTokenManager := mock_auth.NewMockTokenManager(ctrl)
	logger := zap.NewNop()
	svc := service.NewAuthService(logger, mockUserRepo, mockTokenManager)

	email := "test@example.com"
	password := "password"
	role := models.RoleEmployee.String()
	userID := uuid.New()

	t.Run("Successful registration", func(t *testing.T) {
		expectedUser := &models.User{
			ID:           userID,
			Email:        email,
			PasswordHash: "hashed_password",
			Role:         models.RoleEmployee,
		}

		mockUserRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, u *models.User) error {
				assert.Equal(t, email, u.Email)
				assert.NotEqual(t, password, u.PasswordHash)
				assert.Equal(t, models.RoleEmployee, u.Role)

				return nil
			})

		user, err := svc.RegisterUser(context.Background(), email, password, role)

		assert.NoError(t, err)
		assert.Equal(t, expectedUser.Email, user.Email)
		assert.Equal(t, expectedUser.Role, user.Role)
	})

	t.Run("Invalid role", func(t *testing.T) {
		_, err := svc.RegisterUser(context.Background(), email, password, "invalid_role")
		assert.ErrorContains(t, err, domainerrors.ErrInvalidRole.Error())
	})

	t.Run("User already exists", func(t *testing.T) {
		mockUserRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(databaseerrors.ErrUniqueViolation)

		_, err := svc.RegisterUser(context.Background(), email, password, role)
		assert.ErrorIs(t, err, domainerrors.ErrUserExists)
	})

	t.Run("Unexpected repository error", func(t *testing.T) {
		mockUserRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(databaseerrors.ErrUnexpected)

		_, err := svc.RegisterUser(context.Background(), email, password, role)
		assert.ErrorContains(t, err, domainerrors.ErrUnexpected.Error())
	})

	t.Run("Password too long", func(t *testing.T) {
		longPassword := "a" + string(make([]byte, auth.MaxPasswordLength))
		_, err := svc.RegisterUser(context.Background(), email, longPassword, role)
		assert.ErrorIs(t, err, domainerrors.ErrPasswordTooLong)
	})
}

func TestAuthService_AuthenticateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repositories.NewMockIUserRepo(ctrl)
	mockTokenManager := mock_auth.NewMockTokenManager(ctrl)
	logger := zap.NewNop()
	svc := service.NewAuthService(logger, mockUserRepo, mockTokenManager)

	email := "test@example.com"
	password := "password"
	userID := uuid.New()
	testToken := "test_token"

	t.Run("Successful authentication", func(t *testing.T) {
		pwdHash, _ := auth.HashPassword(password)
		user := &models.User{
			ID:           userID,
			Email:        email,
			PasswordHash: pwdHash,
			Role:         models.RoleEmployee,
		}

		mockUserRepo.EXPECT().
			GetByEmail(gomock.Any(), email).
			Return(user, nil)

		mockTokenManager.EXPECT().
			Generate(userID, user.Role.String()).
			Return(testToken, nil)

		token, err := svc.AuthenticateUser(context.Background(), email, password)

		assert.NoError(t, err)
		assert.Equal(t, models.Token(testToken), token)
	})

	t.Run("User not found", func(t *testing.T) {
		mockUserRepo.EXPECT().
			GetByEmail(gomock.Any(), email).
			Return(nil, databaseerrors.ErrNoRows)

		_, err := svc.AuthenticateUser(context.Background(), email, password)
		assert.ErrorIs(t, err, domainerrors.ErrUserNotFound)
	})

	t.Run("Invalid password", func(t *testing.T) {
		pwdHash, _ := auth.HashPassword("wrong_password")
		user := &models.User{
			PasswordHash: pwdHash,
		}

		mockUserRepo.EXPECT().
			GetByEmail(gomock.Any(), email).
			Return(user, nil)

		_, err := svc.AuthenticateUser(context.Background(), email, password)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidCredentials)
	})

	t.Run("Token generation error", func(t *testing.T) {
		pwdHash, _ := auth.HashPassword(password)
		user := &models.User{
			PasswordHash: pwdHash,
			Role:         models.RoleEmployee,
		}

		mockUserRepo.EXPECT().
			GetByEmail(gomock.Any(), email).
			Return(user, nil)

		mockTokenManager.EXPECT().
			Generate(gomock.Any(), gomock.Any()).
			Return("", errors.New("generation error"))

		_, err := svc.AuthenticateUser(context.Background(), email, password)
		assert.ErrorIs(t, err, domainerrors.ErrUnexpected)
	})

	t.Run("Password too long", func(t *testing.T) {
		longPassword := "a" + string(make([]byte, auth.MaxPasswordLength))
		_, err := svc.AuthenticateUser(context.Background(), email, longPassword)
		assert.ErrorIs(t, err, domainerrors.ErrPasswordTooLong)
	})
}

func TestAuthService_DummyLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTokenManager := mock_auth.NewMockTokenManager(ctrl)
	logger := zap.NewNop()
	svc := service.NewAuthService(logger, nil, mockTokenManager)

	role := models.RoleModerator.String()
	testToken := "dummy_token"

	t.Run("Successful dummy login", func(t *testing.T) {
		mockTokenManager.EXPECT().
			Generate(gomock.Any(), role).
			Return(testToken, nil)

		token, err := svc.DummyLogin(context.Background(), role)
		assert.NoError(t, err)
		assert.Equal(t, models.Token(testToken), token)
	})

	t.Run("Invalid role", func(t *testing.T) {
		_, err := svc.DummyLogin(context.Background(), "invalid_role")
		assert.ErrorIs(t, err, domainerrors.ErrInvalidRole)
	})

	t.Run("Token generation failure", func(t *testing.T) {
		mockTokenManager.EXPECT().
			Generate(gomock.Any(), role).
			Return("", errors.New("generation error"))

		_, err := svc.DummyLogin(context.Background(), role)
		assert.ErrorIs(t, err, domainerrors.ErrUnexpected)
	})
}
