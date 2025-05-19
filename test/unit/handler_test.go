package unit

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/l4ndm1nes/Weather-API-Application/internal/handler"
	"github.com/l4ndm1nes/Weather-API-Application/internal/mocks"
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
	"github.com/l4ndm1nes/Weather-API-Application/internal/service"
	"github.com/l4ndm1nes/Weather-API-Application/pkg/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSubscriptionHandler_Subscribe(t *testing.T) {
	tests := []struct {
		name       string
		inputBody  gin.H
		mockSetup  func(svc *mocks.SubscriptionService)
		wantStatus int
	}{
		{
			name: "valid",
			inputBody: gin.H{
				"email":     "test@email.com",
				"city":      "Kyiv",
				"frequency": "daily",
			},
			mockSetup: func(svc *mocks.SubscriptionService) {
				svc.On("Subscribe", mock.Anything).Return(&model.Subscription{
					Email:     "test@email.com",
					City:      "Kyiv",
					Frequency: "daily",
				}, nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "email already subscribed",
			inputBody: gin.H{
				"email":     "dup@email.com",
				"city":      "Kyiv",
				"frequency": "daily",
			},
			mockSetup: func(svc *mocks.SubscriptionService) {
				svc.On("Subscribe", mock.Anything).Return(nil, errors.New("email already subscribed")).Once()
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "invalid input",
			inputBody: gin.H{
				"city": "Kyiv",
			},
			mockSetup:  func(svc *mocks.SubscriptionService) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			subMock := &mocks.SubscriptionService{}
			weatherMock := &mocks.WeatherService{}
			tc.mockSetup(subMock)
			h := handler.NewSubscriptionHandler(subMock, weatherMock)

			r := gin.Default()
			r.POST("/subscribe", h.Subscribe)

			body, _ := json.Marshal(tc.inputBody)
			req := httptest.NewRequest(http.MethodPost, "/subscribe", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestSubscriptionHandler_ConfirmSubscription(t *testing.T) {
	mockRepo := &mocks.SubscriptionRepository{}
	subService := service.NewSubscriptionService(mockRepo, nil) // Без mailer
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

	mockRepo.On("Update", mock.Anything).Return(nil)

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

func TestSubscriptionHandler_GetWeather(t *testing.T) {
	tests := []struct {
		name       string
		queryCity  string
		mockSetup  func(ws *mocks.WeatherService)
		wantStatus int
	}{
		{
			name:      "success",
			queryCity: "Kyiv",
			mockSetup: func(ws *mocks.WeatherService) {
				ws.On("GetWeather", "Kyiv").Return(&model.Weather{
					Temperature: 20,
					Humidity:    60,
					Description: "Sunny",
				}, nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "city not found",
			queryCity: "Atlantis",
			mockSetup: func(ws *mocks.WeatherService) {
				ws.On("GetWeather", "Atlantis").Return(nil, errors.New("City not found")).Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "city missing",
			queryCity:  "",
			mockSetup:  func(ws *mocks.WeatherService) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			subMock := &mocks.SubscriptionService{}
			weatherMock := &mocks.WeatherService{}
			tc.mockSetup(weatherMock)
			h := handler.NewSubscriptionHandler(subMock, weatherMock)

			r := gin.Default()
			r.GET("/weather", h.GetWeather)
			req := httptest.NewRequest(http.MethodGet, "/weather?city="+tc.queryCity, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)
			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestSubscriptionHandler_Unsubscribe(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		mockSetup  func(svc *mocks.SubscriptionService)
		wantStatus int
	}{
		{
			name:  "success",
			token: "5f2a17b1-110c-4881-bc19-41c3edaa0657",
			mockSetup: func(svc *mocks.SubscriptionService) {
				svc.On("Unsubscribe", "5f2a17b1-110c-4881-bc19-41c3edaa0657").Return(nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "token not found",
			token: "dfc16b26-842a-4c8e-b31c-53c6a29360e6",
			mockSetup: func(svc *mocks.SubscriptionService) {
				svc.On("Unsubscribe", "dfc16b26-842a-4c8e-b31c-53c6a29360e6").Return(gorm.ErrRecordNotFound).Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid token",
			token:      "",
			mockSetup:  func(svc *mocks.SubscriptionService) {},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			subMock := &mocks.SubscriptionService{}
			weatherMock := &mocks.WeatherService{}
			tc.mockSetup(subMock)
			h := handler.NewSubscriptionHandler(subMock, weatherMock)

			r := gin.Default()
			r.GET("/unsubscribe/:token", middleware.TokenUUIDRequiredMiddleware("token", "Invalid token"), h.Unsubscribe)
			url := "/unsubscribe/" + tc.token
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)
			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}
