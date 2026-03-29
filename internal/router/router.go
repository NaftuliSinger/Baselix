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

	// API routes
	api := r.Group("/api")
	api.Use(middleware.RequireAPIKey())

	// V1 API endpoints
	v1 := api.Group("/v1")
	v1.GET("/tables", apiHandlers.GetTables)
	v1.GET("/tables/:name", apiHandlers.GetTable)
	v1.POST("/tables/:name", apiHandlers.CreateTable)
	v1.PUT("/tables/:name", apiHandlers.UpdateTable)

	return r
}
