package apiHandlers

import (
	"baselix/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateTable(c *gin.Context) {
	project := middleware.GetAPIProject(c)
	if project == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
}
