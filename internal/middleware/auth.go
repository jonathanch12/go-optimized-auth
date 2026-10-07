package middleware

import (
	"errors"
	"net/http"
	"strings"

	"go-auth-api/internal/auth"
	"go-auth-api/internal/store"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserID = "userID"
	ContextJTI    = "jti"
)

func RequireAuth(sessions store.SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := auth.ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userID, err := sessions.GetUserID(c.Request.Context(), claims.ID)
		if err != nil {
			if errors.Is(err, store.ErrSessionNotFound) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		c.Set(ContextUserID, userID)
		c.Set(ContextJTI, claims.ID)
		c.Next()
	}
}
