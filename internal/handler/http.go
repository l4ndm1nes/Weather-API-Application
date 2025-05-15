package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/l4ndm1nes/Weather-API-Application/internal/app/service"
	"gorm.io/gorm"
	"net/http"
)

type SubscriptionHandler struct {
	SubService     *service.SubscriptionService
	WeatherService *service.WeatherService
}

func NewSubscriptionHandler(subService *service.SubscriptionService, weatherService *service.WeatherService) *SubscriptionHandler {
	return &SubscriptionHandler{
		SubService:     subService,
		WeatherService: weatherService,
	}
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

func (h *SubscriptionHandler) ConfirmSubscription(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(400, gin.H{"error": "Invalid token"})
		return
	}

	err := h.SubService.ConfirmSubscription(token)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(404, gin.H{"error": "Token not found"})
		} else {
			c.JSON(400, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(200, gin.H{"message": "Subscription confirmed successfully"})
}

func (h *SubscriptionHandler) GetWeather(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "city is required"})
		return
	}

	weather, err := h.WeatherService.GetWeather(city)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "City not found"})
		return
	}

	c.JSON(http.StatusOK, weather)
}

func (h *SubscriptionHandler) Unsubscribe(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(400, gin.H{"error": "Invalid token"})
		return
	}
	err := h.SubService.Unsubscribe(token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"error": "Token not found"})
		} else {
			c.JSON(400, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(200, gin.H{"message": "Unsubscribed successfully"})
}

func RegisterRoutes(r *gin.Engine, subHandler *SubscriptionHandler) {
	api := r.Group("/api")
	{
		api.POST("/subscribe", subHandler.Subscribe)
		api.GET("/confirm/:token", subHandler.ConfirmSubscription)
		api.GET("/weather", subHandler.GetWeather)
		api.GET("/unsubscribe/:token", subHandler.Unsubscribe)
	}
}
