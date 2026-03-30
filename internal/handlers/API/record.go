package apiHandlers

import (
	"baselix/internal/db"
	"baselix/internal/middleware"
	"baselix/internal/models"
	"baselix/internal/types"
	"baselix/internal/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateSingleOrMultipleRecords(c *gin.Context) {
	project := middleware.GetAPIProject(c)
	if project == nil {
		utils.ApiError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	tableName := c.Param("name")

	// Read body once; try array first, then single object
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		utils.ApiError(c, http.StatusBadRequest, "failed to read request body")
		return
	}

	var records []map[string]any
	if err := json.Unmarshal(body, &records); err != nil {
		var single map[string]any
		if err := json.Unmarshal(body, &single); err != nil {
			utils.ApiError(c, http.StatusBadRequest, "invalid request body")
			return
		}
		records = []map[string]any{single}
	}

	if len(records) == 0 {
		utils.ApiError(c, http.StatusBadRequest, "no records provided")
		return
	}

	// Infer schema from all records combined, then create/get the table
	schema := utils.InferSchemaFromRecords(records)

	cleanedSchema := utils.RemoveIDFromSchemaMap(schema)
	fields := utils.ConvertSchemaMapToFields(cleanedSchema)

	// Ensure the table exists and get its details
	table, newTableCreated, err := db.ExistsOrCreateTable(c, project.ID, tableName, fields)
	if err != nil {
		utils.ApiError(c, http.StatusInternalServerError, "failed to create or get table")
		return
	}

	// Build a name to Field map for value construction
	fieldMap := make(map[string]*models.Field, len(table.Fields))
	for _, f := range table.Fields {
		fieldMap[f.Name] = f
	}

	// Build Record+Value objects
	modelRecords := make([]*models.Record, len(records))
	for i, data := range records {
		newRecID := uuid.New()
		rec := &models.Record{
			ID:        newRecID,
			ProjectID: project.ID,
			TableID:   table.ID,
		}
		rec.Values = make([]*models.Value, 0, len(data))

		for key, value := range data {
			field, exists := fieldMap[key]
			if !exists {
				utils.ApiError(c, http.StatusBadRequest, fmt.Sprintf("field %q not found in table schema", key))
				return
			}
			// get the type from field and pass into the next func
			val, err := db.NewValue(rec.ID, field.ID, field.Type, value)
			if err != nil {
				utils.ApiError(c, http.StatusInternalServerError, fmt.Sprintf("failed to create value for field %q", key))
				return
			}
			val.Field = field
			rec.Values = append(rec.Values, val)
		}
		modelRecords[i] = rec
	}

	inserted, err := db.InsertRecords(c, modelRecords)
	if err != nil {
		utils.ApiError(c, http.StatusInternalServerError, "failed to insert records")
		return
	}

	result := make([]types.RecordResponse, len(inserted))
	for i, rec := range inserted {
		result[i] = utils.MapRecordModelToRecordResponse(rec)
	}

	if newTableCreated {
		c.JSON(http.StatusCreated, gin.H{
			"message": "Table '" + table.Name + "' created with inferred schema.",
			"records": result,
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"records": result,
	})
}

func GetRecords(c *gin.Context) {
	project := middleware.GetAPIProject(c)
	if project == nil {
		utils.ApiError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	tableName := c.Param("name")

	table, err := db.GetTableByName(c, project.ID, tableName)
	if err != nil {
		utils.ApiError(c, http.StatusNotFound, "table not found")
		return
	}

	records, err := db.GetRecordsByTableID(c, project.ID, table.ID)
	if err != nil {
		utils.ApiError(c, http.StatusInternalServerError, "failed to fetch records, error: "+err.Error())
		return
	}

	result := make([]types.RecordResponse, len(records))
	for i, rec := range records {
		result[i] = utils.MapRecordModelToRecordResponse(rec)
	}

	c.JSON(http.StatusOK, gin.H{
		"records": result,
	})
}

func GetRecordByID(c *gin.Context) {
	project := middleware.GetAPIProject(c)
	if project == nil {
		utils.ApiError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	recordIDStr := c.Param("id")

	recordID, err := uuid.Parse(recordIDStr)
	if err != nil {
		utils.ApiError(c, http.StatusBadRequest, "invalid record ID")
		return
	}

	record, err := db.GetRecordByID(c, project.ID, recordID)
	if err != nil {
		utils.ApiError(c, http.StatusNotFound, "record not found")
		return
	}

	result := utils.MapRecordModelToRecordResponse(record)

	c.JSON(http.StatusOK, gin.H{
		"record": result,
	})
}
