package middleware

import (
	"strings"

	"github.com/MosinFAM/vk-marketplace/internal/auth"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if uid, err := auth.ParseToken(token); err == nil {
				c.Set("user_id", uid)
			}
		}
		c.Next()
	}
}
