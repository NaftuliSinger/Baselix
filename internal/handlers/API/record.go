package apiHandlers

import (
	"baselix/internal/config"
	"baselix/internal/db"
	"baselix/internal/middleware"
	"baselix/internal/models"
	"baselix/internal/types"
	"baselix/internal/utils"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

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

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		utils.ApiError(c, http.StatusBadRequest, "failed to read request body")
		return
	}

	records, err := utils.ParseRecordsBody(body)
	if err != nil {
		utils.ApiError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Verify that the number of records in the request does not exceed the fixed limit - MaxRecordsPerPostRequest from ENV (e.g. 1000)
	if utils.ExceedsRecordFixedLimit(records) {
		utils.ApiError(c, http.StatusBadRequest, fmt.Sprintf("number of records in request body exceeds the maximum allowed limit of %d", config.Cfg.MaxRecordsPerPostRequest))
		return
	}

	/*
		Checking user plan limits:
	*/

	// user plan
	plan := middleware.GetAPIPlan(c)

	if plan == "" {
		utils.ApiError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// records_per_project limit
	limit := config.GetPlanLimit(plan, "records_per_project", 0)

	// get the total records for this project
	totalRecordsInProject, err := db.CountRecordsByProjectID(c, project.ID)

	// get the total count of values across all records in the request and return an error if it exceeds a certain limit (e.g. 10,000) to prevent abuse and protect the system from OOM crashes
	totalValuesInRequest := utils.CountTotalValuesInRecords(records)

	total := totalRecordsInProject + totalValuesInRequest

	// final limit check
	if total > limit {
		utils.ApiError(c, http.StatusPaymentRequired, fmt.Sprintf("plan limits exceeded: total records in project (%d) + total values in request (%d) exceeds your plan limits (%d), please upgrade your plan to increase your limits", totalRecordsInProject, totalValuesInRequest, limit))
		return
	}

	// Infer schema from all records combined, then create/get the table
	schema, err := utils.InferSchemaFromRecords(records)
	if err != nil {
		utils.ApiError(c, http.StatusInternalServerError, "failed to infer schema from records, error: "+err.Error())
		return
	}

	// Convert the inferred schema to Field models, which includes validation
	fields, err := utils.CleanAndConvertPayloadToFieldModels(schema)

	if err != nil {
		var reservedErr *types.ReservedFieldError
		if errors.As(err, &reservedErr) {
			utils.ApiError(c, http.StatusBadRequest, err.Error())
			return
		} else {
			utils.ApiError(c, http.StatusInternalServerError, "failed to convert schema to fields, error: "+err.Error())
			return
		}
	}

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
			val, err := db.NewValue(table.ID, rec.ID, field.ID, field.Type, field.Unique, value)
			if err != nil {
				utils.ApiError(c, http.StatusInternalServerError, fmt.Sprintf("failed to create value for field %q, error: %v", key, err))
				return
			}

			val.Field = field
			val.Table = table
			rec.Values = append(rec.Values, val)
		}
		modelRecords[i] = rec
	}

	inserted, err := db.InsertRecords(c, modelRecords)
	if err != nil {
		var uniqueErr *types.UniqueDuplicateValueError
		if errors.As(err, &uniqueErr) {
			if newTableCreated {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "Table '" + table.Name + "' created with inferred schema.",
					"error":   http.StatusBadRequest,
					"details": uniqueErr.Error(),
				})
				return
			}
			utils.ApiError(c, http.StatusBadRequest, uniqueErr.Error())
			return
		}
		if newTableCreated {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Table '" + table.Name + "' created with inferred schema.",
				"error":   http.StatusInternalServerError,
				"details": "failed to insert records, error: " + err.Error(),
			})
			return
		}
		utils.ApiError(c, http.StatusInternalServerError, "failed to insert records, error: "+err.Error())
		return
	}

	result := make([]types.RecordResponse, len(inserted))
	for i, rec := range inserted {
		rec.Table = table
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

func UpdateSingleOrMultipleRecords(c *gin.Context) {
	project := middleware.GetAPIProject(c)
	if project == nil {
		utils.ApiError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	tableName := c.Param("name")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		utils.ApiError(c, http.StatusBadRequest, "failed to read request body")
		return
	}

	records, err := utils.ParseRecordsBody(body)
	if err != nil {
		utils.ApiError(c, http.StatusBadRequest, err.Error())
		return
	}

	// return an error if created_at or updated_at fields are included in the payload
	for _, rec := range records {
		if _, exists := rec["created_at"]; exists {
			utils.ApiError(c, http.StatusBadRequest, (&types.ReservedFieldError{FieldName: "created_at"}).Error())
			return
		}
		if _, exists := rec["updated_at"]; exists {
			utils.ApiError(c, http.StatusBadRequest, (&types.ReservedFieldError{FieldName: "updated_at"}).Error())
			return
		}
	}

	// convert the body to a map of recordID -> field updates
	updates, err := utils.RecordsToUUIDMap(records)
	if err != nil {
		utils.ApiError(c, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := db.UpdateRecordsByID(c, project.ID, tableName, updates)
	if err != nil {
		var tableNotFoundErr *types.TableNotFoundError
		if errors.As(err, &tableNotFoundErr) {
			utils.ApiError(c, http.StatusNotFound, tableNotFoundErr.Error())
			return
		}

		var recordNotFoundErr *types.RecordNotFoundError
		if errors.As(err, &recordNotFoundErr) {
			utils.ApiError(c, http.StatusNotFound, recordNotFoundErr.Error())
			return
		}
		var uniqueErr *types.UniqueDuplicateValueError
		if errors.As(err, &uniqueErr) {
			utils.ApiError(c, http.StatusBadRequest, uniqueErr.Error())
			return
		}
		utils.ApiError(c, http.StatusInternalServerError, "failed to update records, error: "+err.Error())
		return
	}

	result := make([]types.RecordResponse, len(updated))
	for i, rec := range updated {
		result[i] = utils.MapRecordModelToRecordResponse(rec)
	}

	c.JSON(http.StatusOK, gin.H{
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

	table, err := db.GetTableWithFieldsByName(c, project.ID, tableName)
	if err != nil {
		utils.ApiError(c, http.StatusNotFound, "table not found")
		return
	}

	query := c.Request.URL.Query()

	filterStr := query.Get("filter")
	sortStr := query.Get("sort")

	filters, err := types.ParseFilters(filterStr)
	if err != nil {
		utils.ApiError(c, http.StatusBadRequest, "invalid filter format: "+err.Error())
		return
	}
	sorts, err := types.ParseSorts(sortStr)
	if err != nil {
		utils.ApiError(c, http.StatusBadRequest, "invalid sort format: "+err.Error())
		return
	}

	// Build field map for filter validation and DB query
	fieldMap := make(map[string]*models.Field, len(table.Fields))
	for _, f := range table.Fields {
		fieldMap[f.Name] = f
	}

	// Validate that every non-meta filter field exists in the table schema
	metaFields := map[string]bool{"id": true, "created_at": true, "updated_at": true}
	for _, fl := range filters {
		if !metaFields[fl.Field] {
			if _, ok := fieldMap[fl.Field]; !ok {
				utils.ApiError(c, http.StatusBadRequest, fmt.Sprintf("filter field %q does not exist in table schema", fl.Field))
				return
			}
		}
	}

	records, err := db.GetRecordsByTableID(c, project.ID, table.ID, fieldMap, filters)
	if err != nil {
		utils.ApiError(c, http.StatusInternalServerError, "failed to fetch records, error: "+err.Error())
		return
	}

	result := make([]types.RecordResponse, len(records))
	for i, rec := range records {
		result[i] = utils.MapRecordModelToRecordResponse(rec)
	}

	// Apply sorts (in memory) after fetching from DB
	resultSorted, err := types.ApplySorts(result, sorts)
	if err != nil {
		utils.ApiError(c, http.StatusBadRequest, "invalid sort field: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records": resultSorted,
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
		"records": result,
	})
}

func DeleteRecordByID(c *gin.Context) {
	project := middleware.GetAPIProject(c)
	if project == nil {
		utils.ApiError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	tableName := strings.ToLower(c.Param("name"))

	recordIDStr := c.Param("id")
	recordID, err := uuid.Parse(recordIDStr)
	if err != nil {
		utils.ApiError(c, http.StatusBadRequest, "invalid record ID")
		return
	}

	err = db.DeleteRecordByID(c, project.ID, tableName, recordID)
	if err != nil {
		var recordNotFoundErr *types.RecordNotFoundError
		if errors.As(err, &recordNotFoundErr) {
			utils.ApiError(c, http.StatusNotFound, recordNotFoundErr.Error())
			return
		}

		utils.ApiError(c, http.StatusInternalServerError, "failed to delete record, error: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "record deleted",
	})
}
