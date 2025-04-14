package metrics

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// NewGinMiddleware возвращает мидлварь для GIN.
// Он собирает метрики по запросам и времени ответа.
// Для эндпоинта /metrics метрики не собираются.
func NewGinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		RequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			http.StatusText(status),
		).Inc()

		ResponseTime.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)
	}
}
