package unit

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/l4ndm1nes/Weather-API-Application/internal/handler"
	"github.com/l4ndm1nes/Weather-API-Application/internal/mocks"
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
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
		name           string
		inputBody      gin.H
		mockSetup      func(svc *mocks.SubscriptionService)
		wantStatus     int
		wantInResponse string
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
				}, nil)
			},
			wantStatus:     http.StatusOK,
			wantInResponse: "Subscription successful",
		},
		{
			name: "email already subscribed",
			inputBody: gin.H{
				"email":     "dup@email.com",
				"city":      "Kyiv",
				"frequency": "daily",
			},
			mockSetup: func(svc *mocks.SubscriptionService) {
				svc.On("Subscribe", mock.Anything).Return(nil, errors.New("email already subscribed"))
			},
			wantStatus:     http.StatusConflict,
			wantInResponse: "Email already subscribed",
		},
		{
			name: "invalid input",
			inputBody: gin.H{
				"city": "Kyiv",
			},
			mockSetup:      func(svc *mocks.SubscriptionService) {},
			wantStatus:     http.StatusBadRequest,
			wantInResponse: "required",
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
			assert.Contains(t, w.Body.String(), tc.wantInResponse)
		})
	}
}

func TestSubscriptionHandler_ConfirmSubscription(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		mockSetup      func(svc *mocks.SubscriptionService)
		wantStatus     int
		wantInResponse string
	}{
		{
			name:  "success",
			token: "e7b68c90-d69c-4df6-8a39-5e4e65bcb93c",
			mockSetup: func(svc *mocks.SubscriptionService) {
				svc.On("ConfirmSubscription", "e7b68c90-d69c-4df6-8a39-5e4e65bcb93c").Return(nil)
			},
			wantStatus:     http.StatusOK,
			wantInResponse: "Subscription confirmed",
		},
		{
			name:  "not found",
			token: "bcd6f6b3-7ee5-41e3-9e01-b8a4beac5b0b",
			mockSetup: func(svc *mocks.SubscriptionService) {
				svc.On("ConfirmSubscription", "bcd6f6b3-7ee5-41e3-9e01-b8a4beac5b0b").Return(errors.New("subscription not found"))
			},
			wantStatus:     http.StatusNotFound,
			wantInResponse: "Token not found",
		},
		{
			name:  "already confirmed",
			token: "7ad3be55-32fa-44f2-bf9a-4e8e8e735e0d",
			mockSetup: func(svc *mocks.SubscriptionService) {
				svc.On("ConfirmSubscription", "7ad3be55-32fa-44f2-bf9a-4e8e8e735e0d").Return(errors.New("already confirmed"))
			},
			wantStatus:     http.StatusBadRequest,
			wantInResponse: "already confirmed",
		},
		{
			name:           "missing token",
			token:          "",
			mockSetup:      func(svc *mocks.SubscriptionService) {},
			wantStatus:     http.StatusNotFound,
			wantInResponse: "404 page not found",
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
			r.GET("/confirm/:token", middleware.TokenUUIDRequiredMiddleware("token"), h.ConfirmSubscription)
			url := "/confirm/" + tc.token
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)
			assert.Equal(t, tc.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), tc.wantInResponse)
		})
	}
}

func TestSubscriptionHandler_GetWeather(t *testing.T) {
	tests := []struct {
		name           string
		queryCity      string
		mockSetup      func(ws *mocks.WeatherService)
		wantStatus     int
		wantInResponse string
	}{
		{
			name:      "success",
			queryCity: "Kyiv",
			mockSetup: func(ws *mocks.WeatherService) {
				ws.On("GetWeather", "Kyiv").Return(&model.Weather{
					Temperature: 20,
					Humidity:    60,
					Description: "Sunny",
				}, nil)
			},
			wantStatus:     http.StatusOK,
			wantInResponse: "Sunny",
		},
		{
			name:      "city not found",
			queryCity: "Atlantis",
			mockSetup: func(ws *mocks.WeatherService) {
				ws.On("GetWeather", "Atlantis").Return(nil, errors.New("City not found"))
			},
			wantStatus:     http.StatusNotFound,
			wantInResponse: "City not found",
		},
		{
			name:           "city missing",
			queryCity:      "",
			mockSetup:      func(ws *mocks.WeatherService) {},
			wantStatus:     http.StatusBadRequest,
			wantInResponse: "city is required",
		},
		{
			name:           "city not latin",
			queryCity:      "Київ",
			mockSetup:      func(ws *mocks.WeatherService) {},
			wantStatus:     http.StatusBadRequest,
			wantInResponse: "city must be in Latin letters",
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
			r.GET("/weather",
				middleware.QueryParamRequiredMiddleware("city", middleware.LatinOnlyRegex),
				h.GetWeather,
			)
			req := httptest.NewRequest(http.MethodGet, "/weather?city="+tc.queryCity, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)
			assert.Equal(t, tc.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), tc.wantInResponse)
		})
	}
}

func TestSubscriptionHandler_Unsubscribe(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		mockSetup      func(svc *mocks.SubscriptionService)
		wantStatus     int
		wantInResponse string
	}{
		{
			name:  "success",
			token: "5f2a17b1-110c-4881-bc19-41c3edaa0657",
			mockSetup: func(svc *mocks.SubscriptionService) {
				svc.On("Unsubscribe", "5f2a17b1-110c-4881-bc19-41c3edaa0657").Return(nil)
			},
			wantStatus:     http.StatusOK,
			wantInResponse: "Unsubscribed successfully",
		},
		{
			name:  "token not found",
			token: "dfc16b26-842a-4c8e-b31c-53c6a29360e6",
			mockSetup: func(svc *mocks.SubscriptionService) {
				svc.On("Unsubscribe", "dfc16b26-842a-4c8e-b31c-53c6a29360e6").Return(gorm.ErrRecordNotFound)
			},
			wantStatus:     http.StatusNotFound,
			wantInResponse: "Token not found",
		},
		{
			name:           "invalid token",
			token:          "",
			mockSetup:      func(svc *mocks.SubscriptionService) {},
			wantStatus:     http.StatusNotFound,
			wantInResponse: "404 page not found",
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
			r.GET("/unsubscribe/:token", middleware.TokenUUIDRequiredMiddleware("token"), h.Unsubscribe)
			url := "/unsubscribe/" + tc.token
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)
			assert.Equal(t, tc.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), tc.wantInResponse)
		})
	}
}
