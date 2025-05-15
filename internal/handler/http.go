package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/l4ndm1nes/Weather-API-Application/internal/app/service"
	"net/http"
)

type SubscriptionHandler struct {
	SubService *service.SubscriptionService
}

func NewSubscriptionHandler(subService *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{SubService: subService}
}

type SubscribeRequest struct {
	Email     string `form:"email" json:"email" binding:"required,email"`
	City      string `form:"city" json:"city" binding:"required"`
	Frequency string `form:"frequency" json:"frequency" binding:"required,oneof=hourly daily"`
}

func (h *SubscriptionHandler) Subscribe(c *gin.Context) {
	var req SubscribeRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub, err := h.SubService.Subscribe(req.Email, req.City, req.Frequency)
	if err != nil {
		if err.Error() == "email already subscribed" {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already subscribed"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Subscription successful. Confirmation email sent.",
		"subscription": sub,
	})
}

func RegisterRoutes(r *gin.Engine, subHandler *SubscriptionHandler) {
	api := r.Group("/api")
	{
		api.POST("/subscribe", subHandler.Subscribe)
	}
}
