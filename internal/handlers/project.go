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
		components.Message(
			"Project created successfully!",
			"success",
		))
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
		components.ProjectList(projects),
	)
}
