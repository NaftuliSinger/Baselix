package htmlHandlers

import (
	"fmt"
	"net/http"

	"baselix/internal/config"
	"baselix/internal/db"
	"baselix/internal/middleware"
	"baselix/internal/models"
	"baselix/internal/utils"
	"baselix/internal/views/components"

	"github.com/gin-gonic/gin"
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

	userID := middleware.GetUserID(c)

	/*
		Checking user plan limits:
	*/

	// Get user object
	userObj, err := db.SelectUserByID(c, userID)
	if err != nil {
		utils.RenderHTML(c,
			http.StatusOK,
			components.Message(
				"Error fetching user data: "+err.Error(),
				"error",
			))
		return
	}

	// projects limit
	limit := config.GetPlanLimit(userObj.Plan, "projects", 0)

	// get total projects count for the user
	totalProjects, err := db.CountProjectsByUserID(c, userID)
	if err != nil {
		utils.RenderHTML(c,
			http.StatusOK,
			components.Message(
				"Error fetching projects count: "+err.Error(),
				"error",
			))
		return
	}

	// final limit check
	if totalProjects >= limit {
		utils.RenderHTML(c,
			http.StatusOK,
			components.Message(
				fmt.Sprintf("Project limit reached: you have %d projects which meets or exceeds your plan limit of %d, please upgrade your plan to increase your limits", totalProjects, limit),
				"error",
			))
		return
	}

	// project.ID = uuid.New()
	project.UserID = userID // assign the current user

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
			http.StatusOK,
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

// Delete a project by ID, give a warning about deleting all data and ask for confirmation
func DeleteProjectHTML(c *gin.Context) {
	projectID := c.Param("id")

	if err := db.DeleteProjectByID(c, projectID); err != nil {
		utils.RenderHTML(c,
			http.StatusOK,
			components.Message(
				"Error deleting project: "+err.Error(),
				"error",
			))
		return
	}

	utils.RenderHTML(c,
		http.StatusOK,
		components.Message(
			"Project deleted successfully",
			"success",
		))
}
