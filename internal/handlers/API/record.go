package apiHandlers

import (
	"baselix/internal/db"
	"baselix/internal/middleware"
	"baselix/internal/models"
	"baselix/internal/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
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
		rec := &models.Record{
			ProjectID: project.ID,
			TableID:   table.ID,
		}
		for key, val := range data {
			if key == "id" {
				continue
			}
			field, ok := fieldMap[key]
			if !ok {
				continue
			}
			v := &models.Value{FieldID: field.ID}
			switch field.Type {
			case "string":
				if s, ok := val.(string); ok {
					v.ValueString = s
				} else {
					v.ValueString = fmt.Sprintf("%v", val)
				}
			case "int":
				if n, ok := val.(float64); ok {
					v.ValueInt = int(n)
				}
			case "float":
				if n, ok := val.(float64); ok {
					v.ValueFloat = n
				}
			case "bool":
				if b, ok := val.(bool); ok {
					v.ValueBool = b
				}
			}
			rec.Values = append(rec.Values, v)
		}
		modelRecords[i] = rec
	}

	inserted, err := db.InsertRecords(c, "records", modelRecords)
	if err != nil {
		utils.ApiError(c, http.StatusInternalServerError, "failed to insert records")
		return
	}

	if newTableCreated {
		c.JSON(http.StatusCreated, gin.H{
			"message": "Table '" + table.Name + "' created with inferred schema.",
			"records": inserted,
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"records": inserted,
	})
}
