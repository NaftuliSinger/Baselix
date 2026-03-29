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
	var response []types.TableResponse
	for _, entity := range tables {
		var fields []types.FieldResponse
		for _, field := range entity.Fields {
			fields = append(fields, types.FieldResponse{
				Name: field.Name,
				Type: field.Type,
			})
		}
		response = append(response, types.TableResponse{
			ID:   entity.ID.String(),
			Name: entity.Name,
			Project: types.ProjectResponse{
				ID:          entity.Project.ID.String(),
				Name:        entity.Project.Name,
				Description: entity.Project.Description,
			},
			Fields: fields,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"tables": response,
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
		utils.ApiError(c, http.StatusInternalServerError, "failed to get tables: "+err.Error())
		return
	}

	// map the entities to a response struct that only includes the fields we want to return
	var response types.TableResponse
	var fields []types.FieldResponse
	for _, field := range table.Fields {
		fields = append(fields, types.FieldResponse{
			Name: field.Name,
			Type: field.Type,
		})
	}
	response = types.TableResponse{
		ID:   table.ID.String(),
		Name: table.Name,
		Project: types.ProjectResponse{
			ID:          table.Project.ID.String(),
			Name:        table.Project.Name,
			Description: table.Project.Description,
		},
		Fields: fields,
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
