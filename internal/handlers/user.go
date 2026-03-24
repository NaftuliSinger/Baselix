package handlers

import (
	"net/http"

	"baselix/internal/db"
	"baselix/internal/models"

	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	var users []models.User

	err := db.DB.NewSelect().Model(&users).Scan(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}
