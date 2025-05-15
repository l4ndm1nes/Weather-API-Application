package weatherapi

import (
	"github.com/l4ndm1nes/Weather-API-Application/internal/app/port"
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
)

type WeatherAPIProvider struct {
	apiKey string
}

var _ port.WeatherProvider = (*WeatherAPIProvider)(nil)
