package handler

import (
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
		respondError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	sub, err := h.SubService.Subscribe(ToDomainFromRequest(&req))
	if err != nil {
		if err.Error() == "email already subscribed" {
			respondError(c, http.StatusConflict, "Email already subscribed", err)
		} else {
			respondError(c, http.StatusBadRequest, err.Error(), err)
		}
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{
		"message":      "Subscription successful. Confirmation email sent.",
		"subscription": ToSubscribeResponse(sub),
	})
}

func (h *SubscriptionHandler) ConfirmSubscription(c *gin.Context) {
	token, ok := getStringFromCtx(c, "token")
	if !ok {
		return
	}
	err := h.SubService.ConfirmSubscription(token)
	if err == nil {
		respondSuccess(c, http.StatusOK, gin.H{"message": "Subscription confirmed successfully"})
		return
	}
	handleServiceError(c, err, "Token not found", "already confirmed")
}

func (h *SubscriptionHandler) Unsubscribe(c *gin.Context) {
	token, ok := getStringFromCtx(c, "token")
	if !ok {
		return
	}
	err := h.SubService.Unsubscribe(token)
	if err == nil {
		respondSuccess(c, http.StatusOK, gin.H{"message": "Unsubscribed successfully"})
		return
	}
	handleServiceError(c, err, "Token not found", "")
}

func (h *SubscriptionHandler) GetWeather(c *gin.Context) {
	city, ok := getStringFromCtx(c, "city")
	if !ok {
		return
	}
	weather, err := h.WeatherService.GetWeather(city)
	if err == nil {
		respondSuccess(c, http.StatusOK, gin.H{
			"weather": ToWeatherResponse(weather),
		})
		return
	}
	respondError(c, http.StatusNotFound, "City not found", err)
}

func RegisterRoutes(r *gin.Engine, subHandler *SubscriptionHandler) {
	api := r.Group("/api")
	{
		api.POST("/subscribe",
			middleware.CityLatinOnlyBodyMiddleware(),
			subHandler.Subscribe,
		)
		api.GET("/weather",
			middleware.QueryParamRequiredMiddleware("city", middleware.LatinOnlyRegex),
			subHandler.GetWeather,
		)
		api.GET("/confirm/:token", middleware.TokenUUIDRequiredMiddleware("token"), subHandler.ConfirmSubscription)
		api.GET("/unsubscribe/:token", middleware.TokenUUIDRequiredMiddleware("token"), subHandler.Unsubscribe)
	}
}
