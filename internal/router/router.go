package router

import (
	htmlHandlers "baselix/internal/handlers/HTML"
	webhookHandlers "baselix/internal/handlers/webhook"
	"baselix/internal/middleware"

	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	r := gin.Default()

	// static files
	r.Static("/static", "./static")

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
	// v1 := api.Group("/v1")

	return r
}
