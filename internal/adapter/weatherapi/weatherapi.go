package weatherapi

import (
	"encoding/json"
	"fmt"
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
	"github.com/l4ndm1nes/Weather-API-Application/pkg"
	"go.uber.org/zap"
	"net/http"
)

type WeatherAPIProvider struct {
	apiKey string
}

func NewWeatherAPIProvider(apiKey string) *WeatherAPIProvider {
	return &WeatherAPIProvider{apiKey: apiKey}
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
		pkg.Logger.Error("Failed to make weather API request",
			zap.String("city", city),
			zap.Error(err),
		)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		pkg.Logger.Warn("Non-200 status from weather API",
			zap.String("city", city),
			zap.Int("status_code", resp.StatusCode),
		)
		return nil, fmt.Errorf("failed to get weather: %s", resp.Status)
	}

	var data weatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		pkg.Logger.Error("Failed to decode weather API response",
			zap.String("city", city),
			zap.Error(err),
		)
		return nil, err
	}

	pkg.Logger.Info("Successfully fetched weather",
		zap.String("city", city),
		zap.Float64("temp_c", data.Current.TempC),
		zap.Int("humidity", data.Current.Humidity),
		zap.String("description", data.Current.Condition.Text),
	)

	return &model.Weather{
		Temperature: data.Current.TempC,
		Humidity:    data.Current.Humidity,
		Description: data.Current.Condition.Text,
	}, nil
}
