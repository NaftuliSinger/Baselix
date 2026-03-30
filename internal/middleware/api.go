package middleware

import (
	"baselix/internal/db"
	"baselix/internal/models"
	"baselix/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RequireAPIKey validates the Bearer token in the Authorization header against stored project API key hashes.
func RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			utils.ApiError(c, http.StatusUnauthorized, "missing or invalid Authorization header")
			c.Abort()
			return
		}

		apiKey := strings.TrimPrefix(authHeader, "Bearer ")
		if apiKey == "" {
			utils.ApiError(c, http.StatusUnauthorized, "missing API key")
			c.Abort()
			return
		}

		hashed, err := utils.HashKey(apiKey)
		if err != nil {
			utils.ApiError(c, http.StatusInternalServerError, "internal server error")
			c.Abort()
			return
		}

		// get the project and the user plan in one query
		project, err := db.SelectProjectWithUserByAPIKeyHash(c.Request.Context(), hashed)
		if err != nil {
			utils.ApiError(c, http.StatusUnauthorized, "invalid API key")
			c.Abort()
			return
		}

		// Attach the project and plan to the Gin context for easy access in handlers
		c.Set("project", project)
		c.Set("plan", project.User.Plan)
		c.Next()
	}
}

// GetAPIProject retrieves the authenticated project from the context (set by RequireAPIKey).
func GetAPIProject(c *gin.Context) *models.Project {
	if val, exists := c.Get("project"); exists {
		if project, ok := val.(*models.Project); ok {
			return project
		}
	}
	return nil
}

func GetAPIPlan(c *gin.Context) string {
	if val, exists := c.Get("plan"); exists {
		if plan, ok := val.(string); ok {
			return plan
		}
	}
	return ""
}
