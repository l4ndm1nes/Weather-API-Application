package port

import "github.com/l4ndm1nes/Weather-API-Application/internal/model"

type WeatherProvider interface {
	GetWeather(city string) (*model.Weather, error)
}
