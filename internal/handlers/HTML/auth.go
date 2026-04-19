package htmlHandlers

import (
	"net/http"

	"baselix/internal/config"
	"baselix/internal/middleware"
	"baselix/internal/utils"
	"baselix/internal/views/pages"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthCallback(c *gin.Context) {
	utils.RenderHTML(c, http.StatusOK, pages.Base(config.Cfg.ClerkPublishableKey, pages.AuthCallback()))
}

func AuthStatus(c *gin.Context) {
	publicKeyPEM := []byte(config.Cfg.ClerkPublicKey)
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if _, err := middleware.ValidateSessionCookie(c.Request, pubKey); err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Status(http.StatusNoContent)
}