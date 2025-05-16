package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/l4ndm1nes/Weather-API-Application/internal/adapter/repo"
	"github.com/l4ndm1nes/Weather-API-Application/internal/handler"
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
	"github.com/l4ndm1nes/Weather-API-Application/internal/service"
	"github.com/l4ndm1nes/Weather-API-Application/pkg"
	"github.com/stretchr/testify/assert"
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
	defer pgC.Terminate(ctx)

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

	type testCase struct {
		name         string
		body         map[string]string
		wantCode     int
		wantContains string
		wantInDb     bool
	}

	tests := []testCase{
		{
			name: "valid subscription",
			body: map[string]string{
				"email":     "integration@email.com",
				"city":      "Kyiv",
				"frequency": "daily",
			},
			wantCode:     http.StatusOK,
			wantContains: "Subscription successful",
			wantInDb:     true,
		},
		{
			name: "missing email",
			body: map[string]string{
				"city":      "Kyiv",
				"frequency": "daily",
			},
			wantCode:     http.StatusBadRequest,
			wantContains: "required",
			wantInDb:     false,
		},
		{
			name: "invalid frequency",
			body: map[string]string{
				"email":     "second@email.com",
				"city":      "Kyiv",
				"frequency": "weekly",
			},
			wantCode:     http.StatusBadRequest,
			wantContains: "oneof",
			wantInDb:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tc.body)
			req := httptest.NewRequest("POST", "/api/subscribe", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.wantCode, w.Code)
			assert.Contains(t, w.Body.String(), tc.wantContains)

			if tc.wantInDb {
				var dbSub repo.SubscriptionDB
				err = db.First(&dbSub, "email = ?", tc.body["email"]).Error
				assert.NoError(t, err)
				assert.Equal(t, tc.body["city"], dbSub.City)
				assert.Equal(t, tc.body["frequency"], dbSub.Frequency)
			}
		})
	}
}

func TestConfirmSubscription_Integration(t *testing.T) {
	ctx := context.Background()

	pgC, dsn := setupPostgresContainer(ctx, t)
	defer pgC.Terminate(ctx)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)
	err = db.AutoMigrate(&repo.SubscriptionDB{})
	assert.NoError(t, err)

	subscriptionRepo := repo.NewPostgresRepo(db)
	mailer := &dummyMailer{}
	subService := service.NewSubscriptionService(subscriptionRepo, mailer)
	handlerSub := handler.NewSubscriptionHandler(subService, nil)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/api/confirm/:token", handlerSub.ConfirmSubscription)

	token := "test-token-123"
	sub := &repo.SubscriptionDB{
		Email:        "confirmtest@email.com",
		City:         "Kyiv",
		Frequency:    "daily",
		ConfirmToken: token,
		Confirmed:    false,
	}
	err = db.Create(sub).Error
	assert.NoError(t, err)

	alreadyToken := "already-token"
	alreadySub := &repo.SubscriptionDB{
		Email:        "already@email.com",
		City:         "Lviv",
		Frequency:    "daily",
		ConfirmToken: alreadyToken,
		Confirmed:    true,
	}
	err = db.Create(alreadySub).Error
	assert.NoError(t, err)

	tests := []struct {
		name           string
		token          string
		wantCode       int
		wantInResponse string
		confirmCheck   *string
	}{
		{
			name:           "valid token",
			token:          token,
			wantCode:       http.StatusOK,
			wantInResponse: "Subscription confirmed successfully",
			confirmCheck:   &sub.Email,
		},
		{
			name:           "already confirmed",
			token:          alreadyToken,
			wantCode:       http.StatusBadRequest,
			wantInResponse: "already confirmed",
			confirmCheck:   nil,
		},
		{
			name:           "not found",
			token:          "not-exist-token",
			wantCode:       http.StatusNotFound,
			wantInResponse: "Token not found",
			confirmCheck:   nil,
		},
		{
			name:           "missing token",
			token:          "",
			wantCode:       http.StatusNotFound,
			wantInResponse: "404 page not found",
			confirmCheck:   nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			url := "/api/confirm/"
			if tc.token != "" {
				url += tc.token
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.wantCode, w.Code)
			assert.Contains(t, w.Body.String(), tc.wantInResponse)

			if tc.confirmCheck != nil && tc.wantCode == http.StatusOK {
				var updated repo.SubscriptionDB
				err = db.First(&updated, "email = ?", *tc.confirmCheck).Error
				assert.NoError(t, err)
				assert.True(t, updated.Confirmed)
			}
		})
	}
}

func TestUnsubscribe_Integration(t *testing.T) {
	ctx := context.Background()

	pgC, dsn := setupPostgresContainer(ctx, t)
	defer pgC.Terminate(ctx)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)
	err = db.AutoMigrate(&repo.SubscriptionDB{})
	assert.NoError(t, err)

	subscriptionRepo := repo.NewPostgresRepo(db)
	mailer := &dummyMailer{}
	subService := service.NewSubscriptionService(subscriptionRepo, mailer)
	handlerSub := handler.NewSubscriptionHandler(subService, nil)

	unsubToken := "unsub-token-123"
	sub := &repo.SubscriptionDB{
		Email:            "unsubscribe@email.com",
		City:             "Kyiv",
		Frequency:        "daily",
		UnsubscribeToken: unsubToken,
		Confirmed:        true,
	}
	err = db.Create(sub).Error
	assert.NoError(t, err)

	tests := []struct {
		name           string
		token          string
		wantCode       int
		wantInResponse string
		checkDeleted   bool
	}{
		{
			name:           "success",
			token:          unsubToken,
			wantCode:       http.StatusOK,
			wantInResponse: "Unsubscribed successfully",
			checkDeleted:   true,
		},
		{
			name:           "not found",
			token:          "not-found-token",
			wantCode:       http.StatusNotFound,
			wantInResponse: "Token not found",
			checkDeleted:   false,
		},
		{
			name:           "missing token",
			token:          "",
			wantCode:       http.StatusNotFound,
			wantInResponse: "404 page not found",
			checkDeleted:   false,
		},
	}

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/api/unsubscribe/:token", handlerSub.Unsubscribe)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			url := "/api/unsubscribe/"
			if tc.token != "" {
				url += tc.token
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.wantCode, w.Code)
			assert.Contains(t, w.Body.String(), tc.wantInResponse)

			if tc.checkDeleted {
				var check repo.SubscriptionDB
				err = db.First(&check, "email = ?", "unsubscribe@email.com").Error
				assert.Error(t, err)
				assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
			}
		})
	}
}

func TestGetWeather_Integration(t *testing.T) {
	ctx := context.Background()
	pgC, dsn := setupPostgresContainer(ctx, t)
	defer pgC.Terminate(ctx)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)
	err = db.AutoMigrate(&repo.SubscriptionDB{})
	assert.NoError(t, err)

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
		name           string
		url            string
		wantStatus     int
		wantInResponse string
	}{
		{
			name:           "valid city",
			url:            "/api/weather?city=Kyiv",
			wantStatus:     http.StatusOK,
			wantInResponse: "Clear",
		},
		{
			name:           "city not found",
			url:            "/api/weather?city=Atlantis",
			wantStatus:     http.StatusNotFound,
			wantInResponse: "City not found",
		},
		{
			name:           "city missing",
			url:            "/api/weather",
			wantStatus:     http.StatusBadRequest,
			wantInResponse: "city is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), tc.wantInResponse)
		})
	}
}
