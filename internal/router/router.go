package router

import (
	"baselix/internal/handlers"
	"baselix/internal/views"

	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		views.Index().Render(c.Request.Context(), c.Writer)
	})

	r.GET("/users", handlers.GetUsers)

	return r
}
