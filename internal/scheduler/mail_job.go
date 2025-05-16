package scheduler

import (
	"fmt"
	"github.com/l4ndm1nes/Weather-API-Application/internal/service"
	"github.com/l4ndm1nes/Weather-API-Application/pkg"
	"go.uber.org/zap"
	"os"
	"time"
)

func MailJob(subService *service.SubscriptionService, weatherService *service.WeatherService) error {
	subs, err := subService.GetAllConfirmed()
	if err != nil {
		pkg.Logger.Error("failed to get confirmed subscriptions", zap.Error(err))
		return fmt.Errorf("failed to get confirmed subscriptions: %w", err)
	}

	now := time.Now()

	for _, sub := range subs {
		if sub.Frequency != "hourly" && sub.Frequency != "daily" {
			continue
		}
		if sub.Frequency == "daily" && sub.LastSentAt != nil && now.Sub(*sub.LastSentAt) < 23*time.Hour {
			continue
		}

		weather, err := weatherService.GetWeather(sub.City)
		if err != nil {
			pkg.Logger.Warn("failed to get weather", zap.String("city", sub.City), zap.Error(err))
			continue
		}

		body := fmt.Sprintf(
			"Hello!\n\nWeather in %s:\nTemperature: %.1f°C\nHumidity: %d%%\nDescription: %s\n\nTo unsubscribe: %s/api/unsubscribe/%s",
			sub.City, weather.Temperature, weather.Humidity, weather.Description, os.Getenv("BASE_URL"), sub.UnsubscribeToken,
		)
		if err := subService.SendWeatherUpdate(sub.Email, body); err != nil {
			pkg.Logger.Warn("failed to send email", zap.String("email", sub.Email), zap.Error(err))
			continue
		}

		sub.LastSentAt = &now
		if err := subService.Update(sub); err != nil {
			pkg.Logger.Warn("failed to update last sent time", zap.String("email", sub.Email), zap.Error(err))
			continue
		}
	}
	return nil
}
