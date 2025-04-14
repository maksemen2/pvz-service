//go:build unit
// +build unit

package jwt

import (
	"github.com/maksemen2/pvz-service/internal/pkg/auth"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/maksemen2/pvz-service/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTManager(t *testing.T) {
	cfg := config.AuthConfig{
		JWTSecret:              "test_secret_key_1234567890",
		TokenExpirationSeconds: 3600,
	}

	manager := NewJWTManager(cfg)
	userID := uuid.New()
	role := "moderator"

	t.Run("Generate and Parse valid token", func(t *testing.T) {
		tokenStr, err := manager.Generate(userID, role)
		require.NoError(t, err)

		claims, err := manager.Parse(tokenStr)
		require.NoError(t, err)

		assert.Equal(t, userID, claims.GetUserID())
		assert.Equal(t, role, claims.GetRole())
	})

	t.Run("Parse expired token", func(t *testing.T) {
		now := time.Now()
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwtClaims{
			UserID: userID,
			Role:   role,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
				IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
				NotBefore: jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			},
		})

		tokenStr, err := token.SignedString([]byte(cfg.JWTSecret))
		require.NoError(t, err)

		_, err = manager.Parse(tokenStr)
		assert.ErrorIs(t, err, auth.ErrTokenExpired)
	})

	t.Run("Parse invalid signature", func(t *testing.T) {
		tokenStr, err := manager.Generate(userID, role)
		require.NoError(t, err)

		corruptedToken := tokenStr[:len(tokenStr)-7] + "invalid"
		_, err = manager.Parse(corruptedToken)
		assert.ErrorIs(t, err, auth.ErrMalformedToken)
	})

	t.Run("Parse malformed token", func(t *testing.T) {
		_, err := manager.Parse("invalid.token.string")
		assert.ErrorIs(t, err, auth.ErrMalformedToken)
	})
}
