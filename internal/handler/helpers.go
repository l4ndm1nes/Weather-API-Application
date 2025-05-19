package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/l4ndm1nes/Weather-API-Application/pkg"
	"go.uber.org/zap"
)

func respondError(c *gin.Context, status int, _ string, err error) {
	entry := pkg.Logger.With(zap.Int("status", status))
	if err != nil {
		entry = entry.With(zap.Error(err))
	}
	entry.Warn("error")
	c.Status(status)
}

func respondSuccess(c *gin.Context, status int, payload gin.H) {
	if payload == nil {
		c.Status(status)
		return
	}

	c.JSON(status, payload)
}

func getStringFromCtx(c *gin.Context, key string) (string, bool) {
	val, exists := c.Get(key)
	if !exists {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}
