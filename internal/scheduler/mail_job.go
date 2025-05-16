package scheduler

import (
	"fmt"
	"github.com/l4ndm1nes/Weather-API-Application/internal/service"
	"os"
	"time"
)

func MailJob(subService *service.SubscriptionService, weatherService *service.WeatherService) error {
	subs, err := subService.GetAllConfirmed()
	if err != nil {
		return fmt.Errorf("failed to get confirmed subscriptions: %w", err)
	}

	now := time.Now()

	for _, sub := range subs {
		if sub.Frequency == "hourly" {
		} else if sub.Frequency == "daily" {
			if sub.LastSentAt != nil && now.Sub(*sub.LastSentAt) < 23*time.Hour {
				continue
			}
		} else {
			continue
		}

		weather, err := weatherService.GetWeather(sub.City)
		if err != nil {
			fmt.Printf("failed to get weather for %s: %v\n", sub.City, err)
			continue
		}
		body := fmt.Sprintf(
			"Hello!\n\nWeather in %s:\nTemperature: %.1f°C\nHumidity: %d%%\nDescription: %s\n\nTo unsubscribe: %s/api/unsubscribe/%s",
			sub.City, weather.Temperature, weather.Humidity, weather.Description, os.Getenv("BASE_URL"), sub.UnsubscribeToken,
		)
		if err := subService.SendWeatherUpdate(sub.Email, body); err != nil {
			fmt.Printf("failed to send email to %s: %v\n", sub.Email, err)
			continue
		}

		sub.LastSentAt = &now
		_ = subService.Update(sub)
	}
	return nil
}
