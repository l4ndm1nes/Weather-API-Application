package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/l4ndm1nes/Weather-API-Application/pkg"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
)

func respondError(c *gin.Context, status int, msg string, err error) {
	entry := pkg.Logger.With(zap.Int("status", status))
	if err != nil {
		entry = entry.With(zap.Error(err))
	}
	entry.Warn(msg)
	c.JSON(status, gin.H{"error": msg})
}

func respondSuccess(c *gin.Context, status int, payload gin.H) {
	c.JSON(status, payload)
}

func handleServiceError(c *gin.Context, err error, notFoundMsg, alreadyMsg string) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound) || err.Error() == "subscription not found":
		respondError(c, http.StatusNotFound, notFoundMsg, err)
	case alreadyMsg != "" && err.Error() == alreadyMsg:
		respondError(c, http.StatusBadRequest, alreadyMsg, err)
	default:
		respondError(c, http.StatusBadRequest, err.Error(), err)
	}
}

func getStringFromCtx(c *gin.Context, key string) (string, bool) {
	val, exists := c.Get(key)
	if !exists {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}
