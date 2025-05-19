package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/l4ndm1nes/Weather-API-Application/pkg"
	"go.uber.org/zap"
)

func respondError(c *gin.Context, status int, message string, err error) {
	entry := pkg.Logger.With(zap.Int("status", status), zap.String("message", message))
	if err != nil {
		entry = entry.With(zap.Error(err))
	}
	entry.Warn("error occurred")

	c.Status(status)
}

func respondSuccess(c *gin.Context, status int, payload gin.H) {
	entry := pkg.Logger.With(zap.Int("status", status))
	if payload != nil {
		entry = entry.With(zap.Any("response", payload))
	}
	entry.Info("successful response")
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
