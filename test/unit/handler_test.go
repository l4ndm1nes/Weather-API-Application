package unit

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/l4ndm1nes/Weather-API-Application/internal/handler"
	"github.com/l4ndm1nes/Weather-API-Application/internal/mocks"
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
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
			token: "tok123",
			mockSetup: func(svc *mocks.SubscriptionService) {
				svc.On("ConfirmSubscription", "tok123").Return(nil)
			},
			wantStatus:     http.StatusOK,
			wantInResponse: "Subscription confirmed",
		},
		{
			name:  "not found",
			token: "notfound",
			mockSetup: func(svc *mocks.SubscriptionService) {
				svc.On("ConfirmSubscription", "notfound").Return(errors.New("subscription not found"))
			},
			wantStatus:     http.StatusNotFound,
			wantInResponse: "Token not found",
		},
		{
			name:  "already confirmed",
			token: "already",
			mockSetup: func(svc *mocks.SubscriptionService) {
				svc.On("ConfirmSubscription", "already").Return(errors.New("already confirmed"))
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
			r.GET("/confirm/:token", h.ConfirmSubscription)
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
			name:           "missing city",
			queryCity:      "",
			mockSetup:      func(ws *mocks.WeatherService) {},
			wantStatus:     http.StatusBadRequest,
			wantInResponse: "city is required",
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
			token: "goodtoken",
			mockSetup: func(svc *mocks.SubscriptionService) {
				svc.On("Unsubscribe", "goodtoken").Return(nil)
			},
			wantStatus:     http.StatusOK,
			wantInResponse: "Unsubscribed successfully",
		},
		{
			name:  "token not found",
			token: "notfound",
			mockSetup: func(svc *mocks.SubscriptionService) {
				svc.On("Unsubscribe", "notfound").Return(gorm.ErrRecordNotFound)
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
			r.GET("/unsubscribe/:token", h.Unsubscribe)
			url := "/unsubscribe/" + tc.token
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)
			assert.Equal(t, tc.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), tc.wantInResponse)
		})
	}
}
