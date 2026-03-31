package utils

import "github.com/gin-gonic/gin"

func ApiError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{"error": statusCode, "message": message})
}
