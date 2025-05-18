package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/l4ndm1nes/Weather-API-Application/internal/adapter/mail"
	"github.com/l4ndm1nes/Weather-API-Application/internal/adapter/repo"
	"github.com/l4ndm1nes/Weather-API-Application/internal/adapter/weatherapi"
	"github.com/l4ndm1nes/Weather-API-Application/internal/config"
	"github.com/l4ndm1nes/Weather-API-Application/internal/handler"
	"github.com/l4ndm1nes/Weather-API-Application/internal/scheduler"
	"github.com/l4ndm1nes/Weather-API-Application/internal/service"
	"github.com/l4ndm1nes/Weather-API-Application/pkg"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	pkg.InitLogger()
	defer func() {
		if err := pkg.Logger.Sync(); err != nil {
			fmt.Printf("failed to sync logger: %v\n", err)
		}
	}()
	cfg := config.LoadConfig()

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=require",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		pkg.Logger.Fatal("failed to connect database", zap.Error(err))
	}

	subscriptionRepo := repo.NewPostgresRepo(db)
	smtpMailer := mail.NewSMTPMailer(
		cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom, cfg.BaseURL,
	)
	weatherProvider := weatherapi.NewWeatherAPIProvider(cfg.WeatherAPIKey)
	subService := service.NewSubscriptionService(subscriptionRepo, smtpMailer)
	weatherService := service.NewWeatherService(weatherProvider)

	c := cron.New()
	if _, err := c.AddFunc("0 * * * *", func() {
		pkg.Logger.Info("Starting scheduled weather mail job...")
		if err := scheduler.MailJob(subService, weatherService); err != nil {
			pkg.Logger.Error("Mail job failed", zap.Error(err))
		} else {
			pkg.Logger.Info("Weather mail job completed successfully")
		}
	}); err != nil {
		pkg.Logger.Fatal("failed to add cron job", zap.Error(err))
	}
	c.Start()

	subHandler := handler.NewSubscriptionHandler(subService, weatherService)

	r := gin.Default()

	r.Use(cors.Default())
	r.Static("/static", "./web/static")

	r.GET("/", func(c *gin.Context) {
		c.File("./web/static/index.html")
	})

	handler.RegisterRoutes(r, subHandler)

	pkg.Logger.Info("API server running", zap.String("addr", ":8080"))
	if err := r.Run(":8080"); err != nil {
		pkg.Logger.Fatal("failed to run server", zap.Error(err))
	}
}
