package router

import (
	"baselix/internal/config"
	"baselix/internal/handlers"
	"baselix/internal/middleware"
	"baselix/internal/utils"
	"baselix/internal/views/pages"
	"net/http"

	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		utils.RenderHTML(c, http.StatusOK, pages.Base(config.Cfg.ClerkPublishableKey, pages.Index()))
	})

	r.GET("/sign-in", func(c *gin.Context) {
		utils.RenderHTML(c, http.StatusOK, pages.Base(config.Cfg.ClerkPublishableKey, pages.SignIn()))
	})

	// Protected routes
	auth := r.Group("/")
	auth.Use(middleware.RequireAuth())
	auth.GET("/dashboard", handlers.Dashboard)

	auth.POST("/projects", handlers.CreateProjectHTML)
	auth.GET("/projects", handlers.GetProjectsHTML)

	return r
}
