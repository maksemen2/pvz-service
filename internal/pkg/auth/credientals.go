package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	UserIDKey = "userID"
	RoleKey   = "role"
)

// GetUserIDFromContext принимает контекст gin и возвращает айди пользователя,
// добавленный в него с помощью JWTAuthMiddleware.
// Возвращает айди пользователя и true, если значение было найдено
// и удалось скастить его или uuid.Nil и false в противном случае.
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	rawUserID, exists := c.Get(UserIDKey)
	if !exists {
		return uuid.Nil, false
	}

	userID, ok := rawUserID.(uuid.UUID)
	if !ok {
		return uuid.Nil, false
	}

	if userID == uuid.Nil {
		return uuid.Nil, false
	}

	return userID, true
}

// GetRoleFromContext принимает контекст gin и возвращает роль пользователя,
// добавленную в него с помощью JWTAuthMiddleware.
// Возвращает роль пользователя и true, если значение было найдено
// и удалось скастить его до строки или пустую строку и false в противном случае.
func GetRoleFromContext(c *gin.Context) (string, bool) {
	rawRole, exists := c.Get(RoleKey)
	if !exists {
		return "", false
	}

	role, ok := rawRole.(string)

	if !ok {
		return "", false
	}

	if role == "" {
		return "", false
	}

	return role, true
}
