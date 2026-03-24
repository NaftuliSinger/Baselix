package middleware

import (
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/gin-gonic/gin"
)

// RequireAuth is a Gin middleware that verifies the Clerk session token.
// It checks the __session cookie first, then falls back to the Authorization header.
// Unauthenticated requests are redirected to /sign-in.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionToken, err := c.Cookie("__session")
		if err != nil {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				sessionToken = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				c.Redirect(http.StatusTemporaryRedirect, "/sign-in")
				c.Abort()
				return
			}
		}

		claims, err := jwt.Verify(c.Request.Context(), &jwt.VerifyParams{
			Token: sessionToken,
		})
		if err != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/sign-in")
			c.Abort()
			return
		}

		c.Set("clerk_user_id", claims.Subject)
		c.Next()
	}
}
