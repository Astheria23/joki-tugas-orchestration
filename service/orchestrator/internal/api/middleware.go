package api

import (
	"net/http"
	"strings"

	"github.com/Astheria23/jokiOrchestrator/shared/auth"
	"github.com/Astheria23/jokiOrchestrator/shared/response"
	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware validates the Authorization header with a Bearer JWT token.
func JWTAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "authorization header is required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, http.StatusUnauthorized, "authorization header format must be Bearer {token}")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := auth.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid token: " + err.Error())
			c.Abort()
			return
		}

		// Store user info in context for downstream handlers
		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}
