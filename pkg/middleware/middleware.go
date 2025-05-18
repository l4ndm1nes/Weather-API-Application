package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

var UUIDRegex = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[1-5][a-fA-F0-9]{3}-[89abAB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$`)
var LatinOnlyRegex = regexp.MustCompile(`^[A-Za-z\s-]+$`)

func errorLatinOnly(paramName string, c *gin.Context) {
	msg := paramName + " must be in Latin letters"
	c.JSON(http.StatusBadRequest, gin.H{"error": msg})
	c.Abort()
}

func checkLatinOnly(value, paramName string, c *gin.Context) bool {
	if !LatinOnlyRegex.MatchString(value) {
		errorLatinOnly(paramName, c)
		return false
	}
	return true
}

func QueryParamRequiredMiddleware(paramName string, pattern *regexp.Regexp) gin.HandlerFunc {
	return func(c *gin.Context) {
		value := c.Query(paramName)
		if value == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": paramName + " is required"})
			c.Abort()
			return
		}
		if pattern != nil && !pattern.MatchString(value) {
			if pattern == LatinOnlyRegex {
				errorLatinOnly(paramName, c)
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid value for parameter: " + paramName})
			c.Abort()
			return
		}
		c.Set(paramName, value)
		c.Next()
	}
}

func CityLatinOnlyBodyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			City string `json:"city"`
		}

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			c.Abort()
			return
		}
		if req.City == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "city is required"})
			c.Abort()
			return
		}
		if !checkLatinOnly(req.City, "city", c) {
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		c.Next()
	}
}

func TokenUUIDRequiredMiddleware(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param(paramName)
		if token == "" || !UUIDRegex.MatchString(token) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or missing token (must be UUID)"})
			c.Abort()
			return
		}
		c.Set(paramName, token)
		c.Next()
	}
}
