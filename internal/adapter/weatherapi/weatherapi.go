package weatherapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
)

type WeatherAPIProvider struct {
	apiKey string
}

func NewWeatherAPIProviderFromEnv() *WeatherAPIProvider {
	return &WeatherAPIProvider{
		apiKey: os.Getenv("WEATHER_API_KEY"),
	}
}

type weatherAPIResponse struct {
	Current struct {
		TempC     float64 `json:"temp_c"`
		Humidity  int     `json:"humidity"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
}

func (w *WeatherAPIProvider) GetWeather(city string) (*model.Weather, error) {
	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", w.apiKey, city)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get weather: %s", resp.Status)
	}

	var data weatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &model.Weather{
		Temperature: data.Current.TempC,
		Humidity:    data.Current.Humidity,
		Description: data.Current.Condition.Text,
	}, nil
}
