package apiHandlers

import (
	"baselix/internal/db"
	"baselix/internal/middleware"
	"baselix/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateSingleOrMultipleRecords(c *gin.Context) {
	project := middleware.GetAPIProject(c)
	if project == nil {
		utils.ApiError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Get table name from URL
	tableName := c.Param("name")

	// check request body, it can be either a single record or an array of records
	// Single record comes raw, multiple,are wrapped in “records”
	var records []map[string]any
	if err := c.BindJSON(&records); err != nil {
		// if error, try to bind to single record
		var record map[string]any
		if err := c.BindJSON(&record); err != nil {
			utils.ApiError(c, http.StatusBadRequest, "invalid request body")
			return
		}
		records = append(records, record)
	}

	// in case of new table, we want to infer the schema from the first record, create the table and fields, then insert the records
	schema := utils.InferSchemaFromRecordData(records[0])
	cleanedSchema := utils.RemoveIDFromSchemaMap(schema) // remove id field if it exists, as we will be adding it automatically to the schema for all tables
	fields := utils.ConvertSchemaMapToFields(cleanedSchema)

	table, newTableCreated, err := db.ExistsOrCreateTable(c, project.ID, tableName, fields)
	if err != nil {
		utils.ApiError(c, http.StatusInternalServerError, "failed to create or get table")
		return
	}

	// Build the records and values to the database, return records with their IDs in the response
	// newRecords := make([]models.Record, len(records))
	// for i, recordData := range records {
	// 	newRecords[i] = models.Record{
	// 		ProjectID: project.ID,
	// 		TableName: tableName,
	// 		Data:      recordData,
	// 	}
	// }

	if newTableCreated {
		c.JSON(http.StatusOK, gin.H{
			"message": "Table '" + table.Name + "' created successfully for project '" + project.Name + "' with inferred schema. Records inserted successfully.",
		})

	}
	c.JSON(http.StatusOK, gin.H{
		"schema": schema,
	})
}
