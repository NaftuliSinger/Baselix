package utils

import (
	"baselix/internal/config"

	"github.com/gin-gonic/gin"
)

func ApiError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{"error": statusCode, "message": message})
}

func ExceedsRecordFixedLimit(m []map[string]any) bool {
	if len(m) > config.Cfg.MaxRecordsPerPostRequest {
		return true
	}
	return false
}

func CountTotalValuesInRecords(records []map[string]any) int {
	return len(records)
}
