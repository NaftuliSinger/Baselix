package htmlHandlers

import (
	"net/http"

	"baselix/internal/config"
	"baselix/internal/utils"
	"baselix/internal/views/pages"

	"github.com/gin-gonic/gin"
)

func Index(c *gin.Context) {
	utils.RenderHTML(c, http.StatusOK, pages.Base(config.Cfg.ClerkPublishableKey, pages.Index()))
}

func SignIn(c *gin.Context) {
	utils.RenderHTML(c, http.StatusOK, pages.Base(config.Cfg.ClerkPublishableKey, pages.SignIn()))
}
