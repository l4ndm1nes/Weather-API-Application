package config

import (
	"github.com/l4ndm1nes/Weather-API-Application/pkg"
	"go.uber.org/zap"
	"os"
)

type Config struct {
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	SMTPHost      string
	SMTPPort      string
	SMTPUser      string
	SMTPPass      string
	SMTPFrom      string
	WeatherAPIKey string
	BaseURL       string
}

func LoadConfig() *Config {
	getEnv := func(key, def string) string {
		val := os.Getenv(key)
		if val != "" {
			return val
		}
		if def != "" {
			pkg.Logger.Warn("env var not set, using default", zap.String("env_var", key), zap.String("default", def))
			return def
		}
		pkg.Logger.Fatal("missing required env variable", zap.String("env_var", key))
		return ""
	}

	return &Config{
		DBHost:        getEnv("DB_HOST", ""),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", ""),
		DBPassword:    getEnv("DB_PASSWORD", ""),
		DBName:        getEnv("DB_NAME", ""),
		SMTPHost:      getEnv("SMTP_HOST", ""),
		SMTPPort:      getEnv("SMTP_PORT", "587"),
		SMTPUser:      getEnv("SMTP_USER", ""),
		SMTPPass:      getEnv("SMTP_PASS", ""),
		SMTPFrom:      getEnv("SMTP_FROM", ""),
		WeatherAPIKey: getEnv("WEATHER_API_KEY", ""),
		BaseURL:       getEnv("BASE_URL", "http://localhost:8080"),
	}
}
