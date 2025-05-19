package middleware

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

var UUIDRegex = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[1-5][a-fA-F0-9]{3}-[89abAB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$`)

func abortWithError(c *gin.Context, status int) {
	c.JSON(status, gin.H{})
	c.Abort()
}

func TokenUUIDRequiredMiddleware(paramName string, errMsg string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param(paramName)
		if token == "" || !UUIDRegex.MatchString(token) {
			abortWithError(c, http.StatusBadRequest)
			return
		}
		c.Set(paramName, token)
		c.Next()
	}
}
