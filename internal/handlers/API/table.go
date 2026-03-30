package apiHandlers

import (
	"baselix/internal/db"
	"baselix/internal/middleware"
	"baselix/internal/types"
	"baselix/internal/utils"
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
		fieldsMap := make(map[string]string)
		for _, field := range entity.Fields {
			fieldsMap[field.Name] = field.Type
		}

		tablesResponse = append(tablesResponse, types.TableResponse{
			ID:     entity.ID.String(),
			Name:   entity.Name,
			Fields: fieldsMap,
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

	table, err := db.GetTableWithFields(c, project.ID, tableName)
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

	fieldsMap := make(map[string]string)
	for _, field := range table.Fields {
		fieldsMap[field.Name] = field.Type
	}
	response = types.TableResponse{
		ID:     table.ID.String(),
		Name:   table.Name,
		Fields: fieldsMap,
	}

	c.JSON(http.StatusOK, gin.H{
		"table": response,
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

	schemaClened := utils.RemoveIDFromSchemaMap(schema)

	fields := utils.ConvertSchemaMapToFields(schemaClened)

	newEntity, err := db.CreateTableWithFields(c, project.ID, tableName, fields)

	if err != nil {
		utils.ApiError(c, http.StatusInternalServerError, "failed to create table: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Table '" + tableName + "' created successfully for project '" + project.Name + "'",
		"table_id":   newEntity.ID,
		"table_name": newEntity.Name,
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

	schema := requestBody.Schema

	schemaClened := utils.RemoveIDFromSchemaMap(schema)

	fields := utils.ConvertSchemaMapToFields(schemaClened)

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
