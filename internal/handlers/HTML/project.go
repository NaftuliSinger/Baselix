package htmlHandlers

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

	apiKey, apiHash, _ := utils.GenerateHashedAPIKey()
	project.APIKeyHash = apiHash

	if err := db.InsertProject(c, &project); err != nil {
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
		components.NewProjectResponse(apiKey),
	)
}

func RotateAPIKeyHTML(c *gin.Context) {
	projectID := c.Param("id")

	project, err := db.SelectProjectByID(c, projectID)

	if err != nil {
		utils.RenderHTML(c,
			http.StatusOK,
			components.Message(
				"Error fetching project: "+err.Error(),
				"error",
			))
		return
	}

	apiKey, apiHash, _ := utils.GenerateHashedAPIKey()
	project.APIKeyHash = apiHash

	if err := db.UpdateProjectByID(c, project); err != nil {
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
		components.NewProjectResponse(apiKey),
	)
}

// List all projects for the current user
func GetProjectsHTML(c *gin.Context) {
	userID := middleware.GetUserID(c)

	projects, err := db.SelectProjectsByUserID(c, userID)

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
