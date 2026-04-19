package middleware

import (
	"baselix/internal/config"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrNoSessionCookie = errors.New("no __session cookie")
	ErrInvalidSession  = errors.New("invalid session token")
	ErrInvalidOrigin   = errors.New("invalid session origin")
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
		claims, err := ValidateSessionCookie(c.Request, pubKey)
		if err != nil {
			if errors.Is(err, ErrNoSessionCookie) {
				fmt.Println("No __session cookie found")
				fmt.Println("Error:", err)
				RedirectToSignIn(c)
				return
			}

			RedirectToSignInAndClearSession(c)
			return
		}

		// Store claims in Gin context
		c.Set("claims", claims)
		c.Next()
	}
}

func ValidateSessionCookie(r *http.Request, pubKey interface{}) (*ClerkClaims, error) {
	cookie, err := r.Cookie("__session")
	if err != nil || cookie.Value == "" {
		return nil, ErrNoSessionCookie
	}

	claims := &ClerkClaims{}
	token, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return pubKey, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidSession
	}

	if claims.Azp != config.Cfg.ROrigin {
		return nil, ErrInvalidOrigin
	}

	return claims, nil
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

func RedirectToSignInAndClearSession(c *gin.Context) {
	// clear cookies
	for _, cookie := range c.Request.Cookies() {
		c.SetCookie(cookie.Name, "", -1, "/", "", false, true)
	}

	c.Redirect(http.StatusFound, "/sign-in")
	c.Abort()
}
