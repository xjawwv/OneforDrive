package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/routestorage/backend/internal/repository"
)

func RequirePermission(permissionKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		perms, err := repository.GetUserPermissions(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "permission check failed"})
			c.Abort()
			return
		}

		if !containsString(perms, permissionKey) {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions", "required": permissionKey})
			c.Abort()
			return
		}

		c.Next()
	}
}

func containsString(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}
