package handler

import (
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
)

func ToSubscribeResponse(sub *model.Subscription) *SubscribeResponse {
	if sub == nil {
		return nil
	}
	return &SubscribeResponse{
		ID:        sub.ID,
		Email:     sub.Email,
		City:      sub.City,
		Frequency: sub.Frequency,
		Confirmed: sub.Confirmed,
		CreatedAt: sub.CreatedAt,
		UpdatedAt: sub.UpdatedAt,
	}
}

func ToDomainFromRequest(req *SubscribeRequest) *model.Subscription {
	if req == nil {
		return nil
	}
	return &model.Subscription{
		Email:     req.Email,
		City:      req.City,
		Frequency: req.Frequency,
	}
}

func ToWeatherResponse(weather *model.Weather) *WeatherResponse {
	if weather == nil {
		return nil
	}
	return &WeatherResponse{
		Temperature: weather.Temperature,
		Humidity:    weather.Humidity,
		Description: weather.Description,
	}
}
