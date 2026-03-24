package handlers

import (
	"net/http"

	"baselix/internal/db"
	"baselix/internal/middleware"
	"baselix/internal/models"
	"baselix/internal/utils"
	"baselix/internal/views/components"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handle form submission and create a project
func CreateProjectHTML(c *gin.Context) {
	var project models.Project
	if err := c.ShouldBind(&project); err != nil {
		utils.RenderHTML(c,
			http.StatusOK,
			components.Message(
				"Error binding project data: "+err.Error(),
				"error",
			))
		return
	}

	project.ID = uuid.New()
	project.UserID = middleware.GetUserID(c) // assign the current user
	project.APIKeyHash, _ = utils.GenerateHashedAPIKey()

	if _, err := db.DB.NewInsert().Model(&project).Exec(c); err != nil {
		utils.RenderHTML(c,
			http.StatusOK,
			components.Message(
				"Error creating project: "+err.Error(),
				"error",
			))
		return
	}

	c.Header("HX-Trigger", "projects-updated")

	utils.RenderHTML(c,
		http.StatusOK,
		components.NewProjectResponse(project.APIKeyHash),
	)
}

func RotateAPIKeyHTML(c *gin.Context) {
	projectID := c.Param("id")
	var project models.Project

	if err := db.DB.NewSelect().Model(&project).Where("id = ?", projectID).Scan(c); err != nil {
		utils.RenderHTML(c,
			http.StatusOK,
			components.Message(
				"Error fetching project: "+err.Error(),
				"error",
			))
		return
	}

	project.APIKeyHash, _ = utils.GenerateHashedAPIKey()

	if _, err := db.DB.NewUpdate().Model(&project).Where("id = ?", projectID).Exec(c); err != nil {
		utils.RenderHTML(c,
			http.StatusOK,
			components.Message(
				"Error updating project: "+err.Error(),
				"error",
			))
		return
	}

	utils.RenderHTML(c,
		http.StatusOK,
		components.NewProjectResponse(project.APIKeyHash),
	)
}

// List all projects for the current user
func GetProjectsHTML(c *gin.Context) {
	var projects []models.Project
	userID := middleware.GetUserID(c)

	err := db.DB.NewSelect().Model(&projects).Where("user_id = ?", userID).Scan(c)
	if err != nil {
		utils.RenderHTML(c,
			http.StatusInternalServerError,
			components.Message(
				"Error fetching projects: "+err.Error(),
				"error",
			))
		return
	}

	utils.RenderHTML(
		c,
		http.StatusOK,
		components.ProjectTable(projects),
	)
}
