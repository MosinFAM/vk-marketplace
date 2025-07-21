package middleware

import (
	"time"

	"github.com/MosinFAM/vk-marketplace/internal/logger"
	"github.com/gin-gonic/gin"
)

func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		logger.LogInfo("HTTP Request", map[string]interface{}{
			"method":   method,
			"path":     path,
			"status":   status,
			"duration": duration.String(),
		})
	}
}
