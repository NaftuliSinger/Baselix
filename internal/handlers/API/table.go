package apiHandlers

import (
	"baselix/internal/db"
	"baselix/internal/middleware"
	"baselix/internal/types"
	"baselix/internal/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetTables(c *gin.Context) {
	project := middleware.GetAPIProject(c)
	if project == nil {
		utils.ApiError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	tables, err := db.GetProjectTablesWithFields(c, project.ID)

	if err != nil {
		utils.ApiError(c, http.StatusInternalServerError, "failed to get tables: "+err.Error())
		return
	}

	// map the entities to a response struct that only includes the fields we want to return
	var tablesResponse []types.TableResponse

	for _, entity := range tables {
		fields := make([]types.TableField, 0, len(entity.Fields))
		for _, field := range entity.Fields {
			val := field.Type
			if field.Unique {
				val = field.Type + "_u"
			}
			fields = append(fields, types.TableField{Key: field.Name, Value: val})
		}

		tablesResponse = append(tablesResponse, types.TableResponse{
			Name:   entity.Name,
			Fields: fields,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"tables": tablesResponse,
	})
}

func GetTable(c *gin.Context) {
	project := middleware.GetAPIProject(c)
	if project == nil {
		utils.ApiError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	tableName := c.Param("name")

	table, err := db.GetTableWithFieldsByName(c, project.ID, tableName)
	if err != nil {
		// if no rows are returned, it means the table doesn't exist, so we return a 404
		if err.Error() == "sql: no rows in result set" {
			utils.ApiError(c, http.StatusNotFound, "table not found")
			return
		}

		utils.ApiError(c, http.StatusInternalServerError, "failed to get tables: "+err.Error())
		return
	}

	// map the entities to a response struct that only includes the fields we want to return
	var response types.TableResponse

	fields := make([]types.TableField, 0, len(table.Fields))
	for _, field := range table.Fields {
		val := field.Type
		if field.Unique {
			val = field.Type + "_u"
		}
		fields = append(fields, types.TableField{Key: field.Name, Value: val})
	}
	response = types.TableResponse{
		Name:   table.Name,
		Fields: fields,
	}

	c.JSON(http.StatusOK, gin.H{
		"tables": response,
	})
}

func CreateTable(c *gin.Context) {
	project := middleware.GetAPIProject(c)
	if project == nil {
		utils.ApiError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// get table name from URL param
	tableName := c.Param("name")

	// access the json body of the request
	var requestBody types.SchemaRequestBody

	if err := c.BindJSON(&requestBody); err != nil {
		utils.ApiError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	schema := requestBody.Schema

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
	newTable, err := db.CreateTableWithFields(c, project.ID, tableName, fields)

	if err != nil {
		var tableExistsErr *types.TableAlreadyExistsError
		if errors.As(err, &tableExistsErr) {
			utils.ApiError(c, http.StatusBadRequest, err.Error())
		} else {
			utils.ApiError(c, http.StatusInternalServerError, "failed to create table: "+err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Table '" + tableName + "' created successfully for project '" + project.Name + "'",
		"table_name": newTable.Name,
	})
}

func UpdateTable(c *gin.Context) {
	project := middleware.GetAPIProject(c)
	if project == nil {
		utils.ApiError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// get table name from URL param
	tableName := c.Param("name")

	// access the json body of the request
	var requestBody types.SchemaRequestBody

	if err := c.BindJSON(&requestBody); err != nil {
		utils.ApiError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	missingFlagMessage := "Warning, this route will perform destructive changes (refer to the API documentation for more details). Destructive changes are not allowed unless 'allow_destructive' is set to true in the request body"

	allow_destructive := requestBody.AllowDestructive
	if !allow_destructive {
		utils.ApiError(c, http.StatusBadRequest, missingFlagMessage)
		return
	}

	schema := requestBody.Schema

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

	updatedTable, err := db.UpdateTableFields(c, tableName, fields)

	if err != nil {
		utils.ApiError(c, http.StatusInternalServerError, "failed to update table: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Table '" + tableName + "' updated successfully for project '" + project.Name + "'",
		"table_id":   updatedTable.ID,
		"table_name": updatedTable.Name,
	})
}

func DeleteTable(c *gin.Context) {
	project := middleware.GetAPIProject(c)
	if project == nil {
		utils.ApiError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// get table name from URL param
	tableName := c.Param("name")

	// access the json body of the request
	var requestBody types.SchemaDeleteRequestBody

	missingFlagMessage := "Warning, this route will perform destructive changes (refer to the API documentation for more details). Destructive changes are not allowed unless 'allow_destructive' is set to true in the request body"

	if err := c.BindJSON(&requestBody); err != nil {
		utils.ApiError(c, http.StatusBadRequest, "invalid request body, "+missingFlagMessage)
		return
	}

	allow_destructive := requestBody.AllowDestructive
	if !allow_destructive {
		utils.ApiError(c, http.StatusBadRequest, missingFlagMessage)
		return
	}

	err := db.DeleteTable(c, project.ID, tableName)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			utils.ApiError(c, http.StatusNotFound, "table not found")
			return
		}
		utils.ApiError(c, http.StatusInternalServerError, "failed to delete table: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Table '" + tableName + "' deleted successfully for project '" + project.Name + "'",
	})
}
