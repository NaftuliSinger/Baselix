package middleware

import (
	"baselix/internal/config"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type ClerkClaims struct {
	Sub string `json:"sub"`
	Pla string `json:"pla"`
	Azp string `json:"azp"`
	Sts string `json:"sts"`
	jwt.RegisteredClaims
}

// RequireAuth verifies the __session cookie using Clerk's public key
func RequireAuth() gin.HandlerFunc {
	publicKeyPEM := []byte(config.Cfg.ClerkPublicKey) // In production, use a secure way to manage this key)

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		panic("invalid Clerk public key: " + err.Error())
	}

	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie("__session")
		if err != nil || cookie.Value == "" {
			RedirectToSignIn(c)
			return
		}

		tokenStr := cookie.Value

		claims := &ClerkClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
				return nil, errors.New("unexpected signing method")
			}
			return pubKey, nil
		})
		if err != nil || !token.Valid {
			RedirectToSignIn(c)
			return
		}

		// Optional: check 'azp' claim
		permittedOrigins := []string{"http://localhost:8080"} // your allowed azp
		allowed := false
		for _, o := range permittedOrigins {
			if claims.Azp == o {
				allowed = true
				break
			}
		}
		if !allowed {
			RedirectToSignIn(c)
			return
		}

		// Store claims in Gin context
		c.Set("claims", claims)
		c.Next()
	}
}

// Helpers to access UID and Plan
func GetUserID(c *gin.Context) string {
	if claims, ok := c.Get("claims"); ok {
		if cc, ok := claims.(*ClerkClaims); ok {
			return cc.Sub
		}
	}
	return ""
}

func GetPlan(c *gin.Context) string {
	if claims, ok := c.Get("claims"); ok {
		if cc, ok := claims.(*ClerkClaims); ok {
			parts := strings.SplitN(cc.Pla, ":", 2)
			if len(parts) == 2 {
				return parts[1] // just the plan name
			}
			return cc.Pla // fallback if no colon
		}
	}
	return ""
}

func RedirectToSignIn(c *gin.Context) {
	c.Redirect(http.StatusFound, "/sign-in")
	c.Abort()
}
