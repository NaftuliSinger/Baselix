package utils

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
)

// RenderHTML renders a templ component as an HTML response.
func RenderHTML(c *gin.Context, status int, component templ.Component) {
	c.Status(status)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
	}
}
