package service

import (
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
)

type WeatherProvider interface {
	GetWeather(city string) (*model.Weather, error)
}

type WeatherService struct {
	Provider WeatherProvider
}

func NewWeatherService(provider WeatherProvider) *WeatherService {
	return &WeatherService{Provider: provider}
}

func (ws *WeatherService) GetWeather(city string) (*model.Weather, error) {
	return ws.Provider.GetWeather(city)
}
