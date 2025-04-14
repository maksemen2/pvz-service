package auth

import (
	"errors"
	"github.com/google/uuid"
)

// Claims - интерфейс, описывающий claims для токенов
type Claims interface {
	GetUserID() uuid.UUID
	GetRole() string
}

// TokenManager - интерфейс, описывающий менеджер токенов.
// Он может использоваться для авторизации пользователей
type TokenManager interface {
	Generate(userID uuid.UUID, role string) (string, error) // Generate создает токен из айди пользователя и его роли и возвращает токен в виде строки.
	Parse(token string) (Claims, error)                     // Parse парсит токен и возвращает его Claims
}

var ErrTokenExpired = errors.New("token expired")
var ErrMalformedToken = errors.New("malformed token")
