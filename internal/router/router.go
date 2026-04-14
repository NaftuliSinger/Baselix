package router

import (
	apiHandlers "baselix/internal/handlers/API"
	htmlHandlers "baselix/internal/handlers/HTML"
	webhookHandlers "baselix/internal/handlers/webhook"
	"baselix/internal/middleware"

	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	r := gin.Default()

	// static files
	r.Static("/static", "./static")

	// Health check endpoint
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Public routes
	r.GET("/", htmlHandlers.Index)
	r.GET("/sign-in", htmlHandlers.SignIn)

	// Webhook route (no auth)
	r.POST("/clerk-webhook", webhookHandlers.SvixWebhook)

	// Protected routes
	auth := r.Group("/")
	auth.Use(middleware.RequireAuth())
	auth.GET("/dashboard", htmlHandlers.Dashboard)

	auth.POST("/projects", htmlHandlers.CreateProjectHTML)
	auth.POST("/projects/:id/rotate-key", htmlHandlers.RotateAPIKeyHTML)
	auth.GET("/projects", htmlHandlers.GetProjectsHTML)

	auth.DELETE("/projects/:id", htmlHandlers.DeleteProjectHTML)

	// API routes
	api := r.Group("/api")
	api.Use(middleware.RequireAPIKey())

	// V1 API endpoints
	v1 := api.Group("/v1")

	// Show all tables for a project
	v1.GET("/tables", apiHandlers.GetTables)

	// Record endpoints (primary resource)
	records := v1.Group("/tables/:name")
	{
		// Get records with filtering and sorting
		records.GET("", apiHandlers.GetRecords)

		// Get single record by ID
		records.GET("/:id", apiHandlers.GetRecordByID)

		// Create single or multiple records (upsert-like behavior), inferring schema and creating table if needed

		// Constraint by fixed limit - MaxRecordsPerPostRequest from ENV (e.g. 1000)
		// Constraint by user plan limits
		records.POST("", apiHandlers.CreateSingleOrMultipleRecords)

		// Patch single or multiple records by ID, including partial updates
		records.PATCH("", apiHandlers.UpdateSingleOrMultipleRecords)

		// Delete single record by ID
		records.DELETE("/:id", apiHandlers.DeleteRecordByID)
	}

	// Schema endpoints (table definition)
	schema := v1.Group("/tables/:name/schema")
	{
		schema.GET("", apiHandlers.GetTable)

		// Constraint by user plan limits
		schema.POST("", apiHandlers.CreateTable)
		schema.PUT("", apiHandlers.UpdateTable)
		schema.DELETE("", apiHandlers.DeleteTable)
	}

	return r
}
