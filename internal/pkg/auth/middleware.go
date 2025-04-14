package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	commonerrors "github.com/maksemen2/pvz-service/internal/common/errors"
	"go.uber.org/zap"
)

const authHeaderPrefix = "Bearer "

// NewGinMiddleware возвращает мидлварь для GIN.
// Он проверяет авторизацию (Bearer token).
// В случае, если токен просрочен или невалиден - прерывает дальнейшие выполнения хендлеров.
// В случае, если токен валиден - прокидывает айди пользователя и роль в контекст.
// Может принимать логгер и структуру, имплементирующую интерфейс TokenManager.
func NewGinMiddleware(logger *zap.Logger, tokenManager TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			logger.Debug("unauthorized access")
			c.AbortWithStatusJSON(http.StatusUnauthorized, commonerrors.Unauthorized())

			return
		}

		if !strings.HasPrefix(tokenStr, authHeaderPrefix) {
			logger.Debug("invalid token", zap.String("token", tokenStr))
			c.AbortWithStatusJSON(http.StatusUnauthorized, commonerrors.Unauthorized())

			return
		}

		token := strings.TrimPrefix(tokenStr, authHeaderPrefix)

		claims, err := tokenManager.Parse(token)
		if err != nil {
			logger.Debug("invalid token", zap.String("token", token), zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, commonerrors.Unauthorized())

			return
		}

		// Прокидываем айди и роль в контекст
		c.Set(UserIDKey, claims.GetUserID())
		c.Set(RoleKey, claims.GetRole())

		logger.Debug("access granted")

		c.Next()
	}
}
