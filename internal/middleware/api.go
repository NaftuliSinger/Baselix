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
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}

		apiKey := strings.TrimPrefix(authHeader, "Bearer ")
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing API key"})
			return
		}

		hashed, err := utils.HashKey(apiKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		// get the project and the user plan in one query
		var project models.Project

		err = db.DB.NewSelect().Model(&project).Relation("User").Where("api_key_hash = ?", hashed).Scan(c.Request.Context())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			return
		}

		// get the user plan

		c.Set("project", &project)
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
