package unit

import (
	"errors"
	"testing"

	"github.com/l4ndm1nes/Weather-API-Application/internal/mocks"
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
	"github.com/l4ndm1nes/Weather-API-Application/internal/service"
	"github.com/l4ndm1nes/Weather-API-Application/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func init() {
	pkg.Logger = zap.NewNop()
}

func TestSubscriptionService_Subscribe(t *testing.T) {
	tests := []struct {
		name              string
		findByEmailResult *model.Subscription
		findByEmailErr    error
		createErr         error
		wantErr           bool
		wantErrMessage    string
	}{
		{
			name:              "new user",
			findByEmailResult: nil,
			findByEmailErr:    nil,
			createErr:         nil,
			wantErr:           false,
		},
		{
			name:              "already subscribed",
			findByEmailResult: &model.Subscription{Email: "already@sub.com"},
			findByEmailErr:    nil,
			createErr:         nil,
			wantErr:           true,
			wantErrMessage:    "email already subscribed",
		},
		{
			name:              "db create error",
			findByEmailResult: nil,
			findByEmailErr:    nil,
			createErr:         errors.New("db fail"),
			wantErr:           true,
			wantErrMessage:    "db fail",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mocks.SubscriptionRepository{}
			mailer := &mocks.Mailer{}

			if !tc.wantErr {
				repo.On("FindByEmail", mock.Anything).Return(nil, nil).Once()
				repo.On("Create", mock.Anything).Return(nil).Once()
				repo.On("FindByEmail", mock.Anything).Return(&model.Subscription{
					Email:     "test@unit.com",
					City:      "Kyiv",
					Frequency: "daily",
					Confirmed: false,
				}, nil).Once()
			} else {
				repo.On("FindByEmail", mock.Anything).Return(tc.findByEmailResult, tc.findByEmailErr)
				repo.On("Create", mock.Anything).Return(tc.createErr)
			}

			mailer.On("SendConfirmation", mock.Anything, mock.Anything).Return(nil)

			svc := service.NewSubscriptionService(repo, mailer)

			sub := &model.Subscription{
				Email:     "test@unit.com",
				City:      "Kyiv",
				Frequency: "daily",
			}

			result, err := svc.Subscribe(sub)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrMessage)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, sub.Email, result.Email)
				assert.Equal(t, sub.City, result.City)
				assert.Equal(t, sub.Frequency, result.Frequency)
				assert.False(t, result.Confirmed)
			}
		})
	}
}

func TestSubscriptionService_ConfirmSubscription(t *testing.T) {
	tests := []struct {
		name             string
		getByTokenSub    *model.Subscription
		getByTokenErr    error
		alreadyConfirmed bool
		updateErr        error
		wantErr          bool
		wantErrMessage   string
	}{
		{
			name:             "success",
			getByTokenSub:    &model.Subscription{Confirmed: false},
			getByTokenErr:    nil,
			alreadyConfirmed: false,
			updateErr:        nil,
			wantErr:          false,
		},
		{
			name:             "already confirmed",
			getByTokenSub:    &model.Subscription{Confirmed: true},
			getByTokenErr:    nil,
			alreadyConfirmed: true,
			updateErr:        nil,
			wantErr:          true,
			wantErrMessage:   "already confirmed",
		},
		{
			name:             "not found",
			getByTokenSub:    nil,
			getByTokenErr:    errors.New("not found"),
			alreadyConfirmed: false,
			updateErr:        nil,
			wantErr:          true,
			wantErrMessage:   service.ErrNotFound.Error(),
		},
		{
			name:             "update error",
			getByTokenSub:    &model.Subscription{Confirmed: false},
			getByTokenErr:    nil,
			alreadyConfirmed: false,
			updateErr:        errors.New("update error"),
			wantErr:          true,
			wantErrMessage:   "update error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mocks.SubscriptionRepository{}
			mailer := &mocks.Mailer{}
			svc := service.NewSubscriptionService(repo, mailer)

			repo.On("GetByToken", mock.Anything).Return(tc.getByTokenSub, tc.getByTokenErr)
			if tc.getByTokenErr == nil && !tc.alreadyConfirmed {
				repo.On("Update", mock.Anything).Return(tc.updateErr)
			}

			err := svc.ConfirmSubscription("sometoken")
			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrMessage)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSubscriptionService_Unsubscribe(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{"success", nil, false},
		{"repo error", errors.New("repo fail"), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mocks.SubscriptionRepository{}
			mailer := &mocks.Mailer{}
			svc := service.NewSubscriptionService(repo, mailer)

			repo.On("UnsubscribeByToken", mock.Anything).Return(tc.err)

			err := svc.Unsubscribe("token")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSubscriptionService_GetAllConfirmed(t *testing.T) {
	tests := []struct {
		name     string
		returned []*model.Subscription
		repoErr  error
		wantErr  bool
	}{
		{
			name:     "success",
			returned: []*model.Subscription{{Email: "1@mail.com"}, {Email: "2@mail.com"}},
			repoErr:  nil,
			wantErr:  false,
		},
		{
			name:     "repo error",
			returned: nil,
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mocks.SubscriptionRepository{}
			mailer := &mocks.Mailer{}
			svc := service.NewSubscriptionService(repo, mailer)

			repo.On("GetAllConfirmed").Return(tc.returned, tc.repoErr)

			subs, err := svc.GetAllConfirmed()
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, subs)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.returned, subs)
			}
		})
	}
}

func TestSubscriptionService_Update(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
		wantErr bool
	}{
		{"success", nil, false},
		{"repo error", errors.New("db error"), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mocks.SubscriptionRepository{}
			mailer := &mocks.Mailer{}
			svc := service.NewSubscriptionService(repo, mailer)

			repo.On("Update", mock.Anything).Return(tc.repoErr)

			err := svc.Update(&model.Subscription{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSubscriptionService_SendWeatherUpdate(t *testing.T) {
	tests := []struct {
		name    string
		mailErr error
		wantErr bool
	}{
		{"success", nil, false},
		{"mail error", errors.New("smtp error"), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mocks.SubscriptionRepository{}
			mailer := &mocks.Mailer{}
			svc := service.NewSubscriptionService(repo, mailer)

			mailer.On("SendWeatherUpdate", mock.Anything, "", mock.Anything).Return(tc.mailErr)

			err := svc.SendWeatherUpdate("test@unit.com", "weather info")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
