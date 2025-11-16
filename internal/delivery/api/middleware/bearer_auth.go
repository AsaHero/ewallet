package middleware

import (
	"strings"

	"github.com/AsaHero/e-wallet/internal/delivery/api/apierr"
	"github.com/AsaHero/e-wallet/pkg/security"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the token from the Authorization header.
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			apierr.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		// Expect the header to be "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			apierr.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		// Parse the JWT token.
		claims, err := security.ValidateToken(tokenString)
		if err != nil {
			apierr.Handle(c, err)
			c.Abort()
			return
		}

		if claims.TgUserID == 0 {
			apierr.Forbidden(c, "Invalid token")
			c.Abort()
			return
		}

		// Store claims in context
		c.Set("userID", claims.UserID)
		c.Set("tgUserID", claims.TgUserID)

		c.Next()
	}
}

// GetUserID extracts userID from context
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("userID")
	if !exists {
		return ""
	}
	return userID.(string)
}
