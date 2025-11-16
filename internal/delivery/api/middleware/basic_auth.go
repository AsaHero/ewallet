package middleware

import (
	"strings"

	"github.com/AsaHero/e-wallet/internal/delivery/api/apierr"
	"github.com/AsaHero/e-wallet/pkg/encrypt"
	"github.com/gin-gonic/gin"
)

func AdminAuthorizer(username, password string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the authorization header.
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			apierr.Unauthorized(c, "Authorization header is required")
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.Abort()
			return
		}

		// Decode the provided credentials
		prefix := "Basic "
		if !strings.HasPrefix(authHeader, prefix) {
			apierr.Unauthorized(c, "Invalid authorization header format")
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.Abort()
			return
		}

		encodedCredentials := authHeader[len(prefix):]
		decodedCredentials, err := encrypt.DecodeBase64(encodedCredentials)
		if err != nil {
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.Abort()
			return
		}

		// Check if the decoded credentials are in "username:password" format
		parts := strings.SplitN(string(decodedCredentials), ":", 2)
		if len(parts) != 2 || parts[0] != username || parts[1] != password {
			apierr.Forbidden(c, "Invalid credentials")
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.Abort()
			return
		}

		c.Next()
	}
}
