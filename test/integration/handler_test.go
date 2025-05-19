package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/l4ndm1nes/Weather-API-Application/internal/adapter/repo"
	"github.com/l4ndm1nes/Weather-API-Application/internal/handler"
	"github.com/l4ndm1nes/Weather-API-Application/internal/mocks"
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
	"github.com/l4ndm1nes/Weather-API-Application/internal/service"
	"github.com/l4ndm1nes/Weather-API-Application/pkg"
	"github.com/l4ndm1nes/Weather-API-Application/pkg/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type dummyMailer struct{}

func (d *dummyMailer) SendConfirmation(email, token string) error              { return nil }
func (d *dummyMailer) SendWeatherUpdate(email, city, weatherInfo string) error { return nil }

type dummyWeatherProvider struct{}

func (d *dummyWeatherProvider) GetWeather(city string) (*model.Weather, error) {
	if city == "Kyiv" {
		return &model.Weather{
			Temperature: 21.5,
			Humidity:    56,
			Description: "Clear",
		}, nil
	}
	return nil, fmt.Errorf("city not found")
}

func init() {
	pkg.Logger = zap.NewNop()
}

func setupPostgresContainer(ctx context.Context, t *testing.T) (testcontainers.Container, string) {
	t.Helper()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}
	pgC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)

	host, err := pgC.Host(ctx)
	assert.NoError(t, err)
	port, err := pgC.MappedPort(ctx, "5432")
	assert.NoError(t, err)

	dsn := fmt.Sprintf("host=%s port=%s user=testuser password=testpass dbname=testdb sslmode=disable",
		host, port.Port())
	return pgC, dsn
}

func TestSubscribe_Integration(t *testing.T) {
	ctx := context.Background()
	pgC, dsn := setupPostgresContainer(ctx, t)
	defer func() {
		if err := pgC.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)
	err = db.AutoMigrate(&repo.SubscriptionDB{})
	assert.NoError(t, err)

	subscriptionRepo := repo.NewPostgresRepo(db)
	mailer := &dummyMailer{}
	subService := service.NewSubscriptionService(subscriptionRepo, mailer)
	subHandler := handler.NewSubscriptionHandler(subService, nil)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/api/subscribe", subHandler.Subscribe)

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{
			name: "valid subscription",
			body: map[string]string{
				"email":     "integration@email.com",
				"city":      "Kyiv",
				"frequency": "daily",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "missing email",
			body: map[string]string{
				"city":      "Kyiv",
				"frequency": "daily",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid frequency",
			body: map[string]string{
				"email":     "second@email.com",
				"city":      "Kyiv",
				"frequency": "weekly",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "already subscribed",
			body: map[string]string{
				"email":     "integration@email.com",
				"city":      "Kyiv",
				"frequency": "daily",
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tc.body)
			req := httptest.NewRequest("POST", "/api/subscribe", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestConfirmSubscription_Integration(t *testing.T) {
	mockRepo := &mocks.SubscriptionRepository{}
	mailer := &dummyMailer{}
	subService := service.NewSubscriptionService(mockRepo, mailer)
	subHandler := handler.NewSubscriptionHandler(subService, nil)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/api/confirm/:token", middleware.TokenUUIDRequiredMiddleware("token", "Invalid token"), subHandler.ConfirmSubscription)

	validToken := "550e8400-e29b-41d4-a716-446655440000"
	mockRepo.On("GetByToken", validToken).Return(&model.Subscription{
		Email:        "confirmtest@email.com",
		City:         "Kyiv",
		Frequency:    "daily",
		ConfirmToken: validToken,
		Confirmed:    false,
	}, nil)

	alreadyToken := "123e4567-e89b-12d3-a456-426614174000"
	mockRepo.On("GetByToken", alreadyToken).Return(&model.Subscription{
		Email:        "already@email.com",
		City:         "Lviv",
		Frequency:    "daily",
		ConfirmToken: alreadyToken,
		Confirmed:    true,
	}, nil)

	notFoundToken := "b472a266-d0bf-4ebd-94a8-6a9655cdd8b3"
	mockRepo.On("GetByToken", notFoundToken).Return(nil, gorm.ErrRecordNotFound)

	mockRepo.On("Update", mock.Anything).Return(nil) // Это заглушка, которая будет использоваться, когда вызывается Update

	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{
			name:       "valid token",
			token:      validToken,
			wantStatus: http.StatusOK,
		},
		{
			name:       "not found",
			token:      notFoundToken,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "already confirmed",
			token:      alreadyToken,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid token (not UUID)",
			token:      "not-a-uuid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing token",
			token:      "",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			url := "/api/confirm/" + tc.token
			req := httptest.NewRequest("GET", url, nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestUnsubscribe_Integration(t *testing.T) {
	ctx := context.Background()
	pgC, dsn := setupPostgresContainer(ctx, t)
	defer func() {
		if err := pgC.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)
	assert.NoError(t, db.AutoMigrate(&repo.SubscriptionDB{}))

	subscriptionRepo := repo.NewPostgresRepo(db)
	mailer := &dummyMailer{}
	subService := service.NewSubscriptionService(subscriptionRepo, mailer)
	subHandler := handler.NewSubscriptionHandler(subService, nil)

	unsubToken := "ae7b31ab-7b5b-4be0-8f89-7e0a9c872f0d"
	sub := &repo.SubscriptionDB{
		Email:            "unsubscribe@email.com",
		City:             "Kyiv",
		Frequency:        "daily",
		UnsubscribeToken: unsubToken,
		Confirmed:        true,
	}
	assert.NoError(t, db.Create(sub).Error)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/api/unsubscribe/:token", middleware.TokenUUIDRequiredMiddleware("token", "Invalid token"), subHandler.Unsubscribe)

	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{
			name:       "success",
			token:      unsubToken,
			wantStatus: http.StatusOK,
		},
		{
			name:       "token not found",
			token:      "dfc16b26-842a-4c8e-b31c-53c6a29360e6",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid token",
			token:      "",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			url := "/api/unsubscribe/" + tc.token
			req := httptest.NewRequest("GET", url, nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestGetWeather_Integration(t *testing.T) {
	ctx := context.Background()
	pgC, dsn := setupPostgresContainer(ctx, t)
	defer func() {
		if err := pgC.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)
	assert.NoError(t, db.AutoMigrate(&repo.SubscriptionDB{}))

	subscriptionRepo := repo.NewPostgresRepo(db)
	mailer := &dummyMailer{}
	weatherProvider := &dummyWeatherProvider{}
	subService := service.NewSubscriptionService(subscriptionRepo, mailer)
	weatherService := service.NewWeatherService(weatherProvider)
	subHandler := handler.NewSubscriptionHandler(subService, weatherService)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/api/weather", subHandler.GetWeather)

	tests := []struct {
		name       string
		queryCity  string
		wantStatus int
	}{
		{
			name:       "success",
			queryCity:  "Kyiv",
			wantStatus: http.StatusOK,
		},
		{
			name:       "city not found",
			queryCity:  "Atlantis",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "city missing",
			queryCity:  "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/weather?city="+tc.queryCity, nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}
