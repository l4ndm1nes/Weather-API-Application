package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
	"github.com/l4ndm1nes/Weather-API-Application/pkg"
	"go.uber.org/zap"
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

func respondError(c *gin.Context, status int, msg string, err error) {
	pkg.Logger.Warn(msg, zap.Error(err))
	c.JSON(status, gin.H{"error": msg})
}

func (h *SubscriptionHandler) Subscribe(c *gin.Context) {
	var req SubscribeRequest
	if err := c.ShouldBind(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	sub, err := h.SubService.Subscribe(ToDomainFromRequest(&req))
	if err != nil {
		if err.Error() == "email already subscribed" {
			respondError(c, http.StatusConflict, "Email already subscribed", err)
			return
		}
		respondError(c, http.StatusBadRequest, err.Error(), err)
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
		respondError(c, http.StatusBadRequest, "Invalid token", nil)
		return
	}

	err := h.SubService.ConfirmSubscription(token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || err.Error() == "subscription not found" {
			respondError(c, http.StatusNotFound, "Token not found", err)
			return
		}
		respondError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription confirmed successfully"})
}

func (h *SubscriptionHandler) GetWeather(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		respondError(c, http.StatusBadRequest, "city is required", nil)
		return
	}

	weather, err := h.WeatherService.GetWeather(city)
	if err != nil {
		respondError(c, http.StatusNotFound, "City not found", err)
		return
	}

	c.JSON(http.StatusOK, ToWeatherResponse(weather))
}

func (h *SubscriptionHandler) Unsubscribe(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		respondError(c, http.StatusBadRequest, "Invalid token", nil)
		return
	}
	err := h.SubService.Unsubscribe(token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(c, http.StatusNotFound, "Token not found", err)
			return
		}
		respondError(c, http.StatusBadRequest, err.Error(), err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Unsubscribed successfully"})
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
