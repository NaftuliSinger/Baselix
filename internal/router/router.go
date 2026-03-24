package router

import (
	"baselix/internal/utils"
	"baselix/internal/views/pages"
	"net/http"

	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		utils.RenderHTML(c, http.StatusOK, pages.Base(pages.Index()))
	})

	r.GET("/dashboard", func(c *gin.Context) {
		utils.RenderHTML(c, http.StatusOK, pages.Base(pages.Dashboard()))
	})

	// r.GET("/users", handlers.GetUsers)

	return r
}
