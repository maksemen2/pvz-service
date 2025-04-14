package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/maksemen2/pvz-service/config"
	"github.com/maksemen2/pvz-service/internal/pkg/auth"
)

// jwtClaims - имплементация auth.TokenManager для jwt-токена.
type jwtClaims struct {
	UserID uuid.UUID `json:"userID"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

// GetUserID - геттер для ID пользователя.
func (c *jwtClaims) GetUserID() uuid.UUID {
	return c.UserID
}

// GetRole - геттер для роли пользователя.
func (c *jwtClaims) GetRole() string {
	return c.Role
}

// Valid проверяет, просрочен ли токен.
// Если да - возвращает ошибку.
func (c *jwtClaims) Valid() error {
	return nil
}

// jwtManager - реализация менеджера токенов на основе JWT.
type jwtManager struct {
	secretKey []byte
	duration  time.Duration
}

// Generate создает новый JWT токен на основе данных о пользователе и длительности.
// Подписывает токен секретным ключом с помощью алгоритма HS256
// и возвращает его в виде строки.
func (m *jwtManager) Generate(userID uuid.UUID, role string) (string, error) {
	currentTime := time.Now()
	eat := currentTime.Add(m.duration)

	claims := &jwtClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(eat),
			IssuedAt:  jwt.NewNumericDate(currentTime),
			NotBefore: jwt.NewNumericDate(currentTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(m.secretKey)
}

// Parse парсит токен и возвращает его claims.
// Если токен невалидный или просрочен,
// возвращает соответствующую ошибку.
func (m *jwtManager) Parse(tokenString string) (auth.Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwtClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, auth.ErrMalformedToken
			}

			return m.secretKey, nil
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenMalformed), errors.Is(err, jwt.ErrSignatureInvalid):
			return nil, auth.ErrMalformedToken
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, auth.ErrTokenExpired
		}

		return nil, err
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		return nil, auth.ErrMalformedToken
	}

	return claims, nil
}

// NewJWTManager - конструктор для создания нового менеджера токенов на основе JWT.
func NewJWTManager(config config.AuthConfig) auth.TokenManager {

	return &jwtManager{
		secretKey: []byte(config.JWTSecret),
		duration:  time.Duration(config.TokenExpirationSeconds) * time.Second,
	}
}
