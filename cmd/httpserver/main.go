package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/l4ndm1nes/Weather-API-Application/internal/adapter/repo"
	"github.com/l4ndm1nes/Weather-API-Application/internal/app/service"
	"github.com/l4ndm1nes/Weather-API-Application/internal/handler"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

func main() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	subscriptionRepo := repo.NewPostgresRepo(db)
	subService := service.NewSubscriptionService(subscriptionRepo)
	subHandler := handler.NewSubscriptionHandler(subService)

	r := gin.Default()
	handler.RegisterRoutes(r, subHandler)

	fmt.Println("API server running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
