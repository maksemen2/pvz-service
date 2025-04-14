package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	domainerrors "github.com/maksemen2/pvz-service/internal/domain/errors"
	"github.com/maksemen2/pvz-service/internal/domain/models"
	"github.com/maksemen2/pvz-service/internal/domain/repositories"
	"github.com/maksemen2/pvz-service/internal/pkg/auth"
	databaseerrors "github.com/maksemen2/pvz-service/internal/repository/errors"
	"go.uber.org/zap"
)

// AuthService - интерфейс для использования функционала аутентификации и авторизации пользователей.
type AuthService interface {
	RegisterUser(ctx context.Context, email, password, role string) (*models.User, error) // Регистрирует пользователя, создает запись о нем в базе данных и возвращает его доменную модель.
	AuthenticateUser(ctx context.Context, email, password string) (models.Token, error)   // Аутентифицирует пользователя, проверяет его логин и пароль, создает токен доступа и возвращает его.
	DummyLogin(ctx context.Context, role string) (models.Token, error)                    // Создает токен доступа для тестирования, возвращает его.
}

// authServiceImpl реализует интерфейс AuthService.
type authServiceImpl struct {
	logger       *zap.Logger
	userRepo     repositories.IUserRepo // Репозиторий пользователей
	tokenManager auth.TokenManager      // Может принимать любой менеджер токенов, реализующий интерфейс TokenManager
}

// NewAuthService создает новый экземпляр authServiceImpl.
// Принимает логгер, репозиторий пользователей и менеджер токенов.
func NewAuthService(logger *zap.Logger, userRepo repositories.IUserRepo, tokenManager auth.TokenManager) AuthService {
	return &authServiceImpl{
		logger:       logger,
		userRepo:     userRepo,
		tokenManager: tokenManager,
	}
}

// RegisterUser регистрирует нового пользователя.
// Возвращает ошибку, если не удалось захешировать пароль,
// если пользователь с таким Email уже существует
// или если возникла непредвиденная ошибка базы данных.
func (a *authServiceImpl) RegisterUser(ctx context.Context, email, password string, role string) (*models.User, error) {
	roleType := models.RoleType(role)

	if !roleType.Valid() {
		return nil, fmt.Errorf("%w: %s", domainerrors.ErrInvalidRole, role)
	}

	// Валидацию почты можно не проводить,
	// так как она проводится на транспортном уровне и запрос с неверной почтой не будет допущен до обработки

	// Надо обработать ограничение bcrypt.
	// Нас интересует количество байт, а не сама длина строки,
	// поэтому используем функцию len
	if len(password) > auth.MaxPasswordLength {
		return nil, fmt.Errorf("%w: %d", domainerrors.ErrPasswordTooLong, auth.MaxPasswordLength)
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		a.logger.Error("failed to hash password", zap.Error(err)) // сам пароль логировать нельзя
		return nil, fmt.Errorf("%w: %v", domainerrors.ErrUnexpected, err)
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		Role:         roleType,
	}

	err = a.userRepo.Create(ctx, user)

	if err != nil {
		switch {
		case errors.Is(err, databaseerrors.ErrUniqueViolation):
			a.logger.Debug("user already exists", zap.String("email", email), zap.Error(err)) // Это ошибка на стороне пользователя, поэтому логируем ее с уровнем Debug
			return nil, domainerrors.ErrUserExists
		case errors.Is(err, databaseerrors.ErrUnexpected):
			return nil, fmt.Errorf("%w: %v", domainerrors.ErrUnexpected, err)
		}
	}

	return user, nil
}

// AuthenticateUser аутентифицирует пользователя по его логину и паролю.
// Возвращает токен доступа, если аутентификация прошла успешно,
// или ошибку, если пользователь не найден, пароль неверный
// или возникла непредвиденная ошибка базы данных.
func (a *authServiceImpl) AuthenticateUser(ctx context.Context, email, password string) (models.Token, error) {
	// Даже не будем обрабатывать запрос,
	// опять же ограничение bcrypt,
	// при валидации пароля всегда получим false
	if len(password) > auth.MaxPasswordLength {
		return "", fmt.Errorf("%w: %d", domainerrors.ErrPasswordTooLong, auth.MaxPasswordLength)
	}

	user, err := a.userRepo.GetByEmail(ctx, email)

	if err != nil {
		switch {
		case errors.Is(err, databaseerrors.ErrNoRows):
			a.logger.Debug("user does not exist", zap.String("email", email))
			return "", domainerrors.ErrUserNotFound
		case errors.Is(err, databaseerrors.ErrUnexpected):
			return "", domainerrors.ErrUnexpected
		}

		return "", err
	}

	if !auth.ComparePassword(password, user.PasswordHash) {
		a.logger.Debug("invalid password", zap.String("email", email))
		return "", domainerrors.ErrInvalidCredentials
	}

	token, err := a.tokenManager.Generate(user.ID, user.Role.String())

	if err != nil {
		a.logger.Error("failed to generate token", zap.Error(err))
		return "", fmt.Errorf("%w: %v", domainerrors.ErrUnexpected, err)
	}

	a.logger.Debug("successfully authenticated user", zap.String("email", email))

	return models.Token(token), nil
}

// DummyLogin создает токен доступа для тестирования.
// Возвращает его, если токен был успешно сгенерирован,
func (a *authServiceImpl) DummyLogin(ctx context.Context, role string) (models.Token, error) {
	roleType := models.RoleType(role)
	if !roleType.Valid() {
		a.logger.Debug("invalid role", zap.String("role", role))
		return "", domainerrors.ErrInvalidRole
	}

	userID := uuid.New() // Генерируем новый UUID для тестового пользователя, чтобы положить его в токен

	token, err := a.tokenManager.Generate(userID, role)
	if err != nil {
		a.logger.Error("failed to generate token", zap.Error(err))
		return "", fmt.Errorf("%w: %v", domainerrors.ErrUnexpected, err)
	}

	return models.Token(token), nil
}
