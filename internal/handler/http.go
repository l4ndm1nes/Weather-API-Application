package handler

import (
	"errors"
	"gorm.io/gorm"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
	"github.com/l4ndm1nes/Weather-API-Application/pkg/middleware"
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
		respondError(c, http.StatusBadRequest, "Invalid input", err)
		return
	}
	_, err := h.SubService.Subscribe(ToDomainFromRequest(&req))
	if err != nil {
		if err.Error() == "email already subscribed" {
			respondError(c, http.StatusConflict, "Email already subscribed", err)
		} else {
			respondError(c, http.StatusBadRequest, "Invalid input", err)
		}
		return
	}

	respondSuccess(c, http.StatusOK, nil)
}

func (h *SubscriptionHandler) ConfirmSubscription(c *gin.Context) {
	token, ok := getStringFromCtx(c, "token")
	if !ok {
		respondError(c, http.StatusBadRequest, "Invalid token", nil)
		return
	}

	err := h.SubService.ConfirmSubscription(token)
	if err != nil {
		if err.Error() == "already confirmed" {
			respondError(c, http.StatusBadRequest, "Subscription already confirmed", nil)
		} else if err.Error() == "subscription not found" {
			respondError(c, http.StatusNotFound, "Token not found", err)
		} else {
			respondError(c, http.StatusBadRequest, "Error confirming subscription", err)
		}
		return
	}

	respondSuccess(c, http.StatusOK, nil)
}

func (h *SubscriptionHandler) Unsubscribe(c *gin.Context) {
	token, ok := getStringFromCtx(c, "token")
	if !ok {
		respondError(c, http.StatusBadRequest, "Invalid token", nil)
		return
	}

	err := h.SubService.Unsubscribe(token)
	if err == nil {
		respondSuccess(c, http.StatusOK, nil)
		return
	}

	if errors.Is(err, gorm.ErrRecordNotFound) || err.Error() == "subscription not found" {
		respondError(c, http.StatusNotFound, "Token not found", err)
		return
	}

	respondError(c, http.StatusBadRequest, "Error unsubscribing", err)
}

func (h *SubscriptionHandler) GetWeather(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		respondError(c, http.StatusBadRequest, "Invalid request", nil)
		return
	}
	weather, err := h.WeatherService.GetWeather(city)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"temperature": weather.Temperature,
			"humidity":    weather.Humidity,
			"description": weather.Description,
		})
		return
	}
	respondError(c, http.StatusNotFound, "City not found", err)
}

func RegisterRoutes(r *gin.Engine, subHandler *SubscriptionHandler) {
	api := r.Group("/api")
	{
		api.POST("/subscribe", subHandler.Subscribe)
		api.GET("/weather", subHandler.GetWeather)
		api.GET("/confirm/:token",
			middleware.TokenUUIDRequiredMiddleware("token", "Invalid token"),
			subHandler.ConfirmSubscription,
		)
		api.GET("/unsubscribe/:token",
			middleware.TokenUUIDRequiredMiddleware("token", "Invalid token"),
			subHandler.Unsubscribe,
		)
	}
}
