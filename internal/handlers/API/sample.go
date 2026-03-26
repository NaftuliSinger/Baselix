package apiHandlers

import (
	"baselix/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SampleAPI(c *gin.Context) {
	project := middleware.GetAPIProject(c)
	plan := middleware.GetAPIPlan(c)
	c.JSON(http.StatusOK, gin.H{"message": "Hello from protected API!", "project_id": project.ID, "plan": plan})
}
