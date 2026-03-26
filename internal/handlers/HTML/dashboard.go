package htmlHandlers

import (
	"net/http"

	"baselix/internal/config"
	"baselix/internal/middleware"
	"baselix/internal/utils"
	"baselix/internal/views/pages"

	"github.com/gin-gonic/gin"
)

func Dashboard(c *gin.Context) {
	uid := middleware.GetUserID(c)
	if uid == "" {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	plan := middleware.GetPlan(c) // plan may be empty if not set

	utils.RenderHTML(
		c,
		http.StatusOK,
		pages.Base(
			config.Cfg.ClerkPublishableKey,
			pages.Dashboard(
				uid,
				plan,
			),
		),
	)
}
