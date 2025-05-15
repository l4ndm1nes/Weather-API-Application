package service

import (
	"github.com/l4ndm1nes/Weather-API-Application/internal/app/port"
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
)

type WeatherService struct {
	Provider port.WeatherProvider
}

func NewWeatherService(provider port.WeatherProvider) *WeatherService {
	return &WeatherService{Provider: provider}
}

func (ws *WeatherService) GetWeather(city string) (*model.Weather, error) {
	return ws.Provider.GetWeather(city)
}
