package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
	"gorm.io/gorm"
	"net/http"
)

type SubscriptionService interface {
	Subscribe(sub *model.Subscription) (*model.Subscription, error)
	ConfirmSubscription(token string) error
	Unsubscribe(token string) error
}

type WeatherService interface {
	GetWeather(city string) (*model.Weather, error)
}

type SubscriptionHandler struct {
	SubService     SubscriptionService
	WeatherService WeatherService
}

func NewSubscriptionHandler(subService SubscriptionService, weatherService WeatherService) *SubscriptionHandler {
	return &SubscriptionHandler{
		SubService:     subService,
		WeatherService: weatherService,
	}
}

func (h *SubscriptionHandler) Subscribe(c *gin.Context) {
	var req SubscribeRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub, err := h.SubService.Subscribe(ToDomainFromRequest(&req))
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
		"subscription": ToSubscribeResponse(sub),
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
		if errors.Is(err, gorm.ErrRecordNotFound) || err.Error() == "subscription not found" {
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

	c.JSON(http.StatusOK, ToWeatherResponse(weather))
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
