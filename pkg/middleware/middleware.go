package middleware

import (
	"go.uber.org/zap"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/l4ndm1nes/Weather-API-Application/pkg"
)

var UUIDRegex = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[1-5][a-fA-F0-9]{3}-[89abAB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$`)

func abortWithError(c *gin.Context, status int) {
	entry := pkg.Logger.With(zap.Int("status", status))
	entry.Warn("invalid token format")

	c.Status(status)
	c.Abort()
}

func TokenUUIDRequiredMiddleware(paramName string, errMsg string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param(paramName)
		entry := pkg.Logger.With(zap.String("token", token))

		if token == "" || !UUIDRegex.MatchString(token) {
			entry.Warn("invalid or missing token")
			abortWithError(c, http.StatusBadRequest)
			return
		}

		entry.Info("valid token")
		c.Set(paramName, token)
		c.Next()
	}
}
