//go:build unit
// +build unit

package auth_test

import (
	"testing"

	"github.com/maksemen2/pvz-service/internal/pkg/auth"
	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		password := "best_password_ever"

		hash, err := auth.HashPassword(password)

		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
	})

	t.Run("Password Too Long", func(t *testing.T) {
		password := "a" + string(make([]byte, auth.MaxPasswordLength))

		hash, err := auth.HashPassword(password)

		assert.Error(t, err)
		assert.Empty(t, hash)
	})
}

func TestComparePassword(t *testing.T) {
	validPassword := "correct_password"
	validHash, _ := auth.HashPassword(validPassword)

	t.Run("Valid password", func(t *testing.T) {
		assert.True(t, auth.ComparePassword(validPassword, validHash))
	})

	t.Run("Wrong password", func(t *testing.T) {
		assert.False(t, auth.ComparePassword("wrong_password", validHash))
	})

	t.Run("Invalid hash format", func(t *testing.T) {
		assert.False(t, auth.ComparePassword(validPassword, "invalid_hash"))
	})

	t.Run("Empty hash", func(t *testing.T) {
		assert.False(t, auth.ComparePassword(validPassword, ""))
	})

	t.Run("Password too long", func(t *testing.T) {
		longPassword := "a" + string(make([]byte, auth.MaxPasswordLength))
		assert.False(t, auth.ComparePassword(longPassword, validHash))
	})
}
