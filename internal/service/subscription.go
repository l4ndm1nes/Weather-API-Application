package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
	"github.com/l4ndm1nes/Weather-API-Application/pkg"
	"go.uber.org/zap"
)

var ErrNotFound = errors.New("subscription not found")

type SubscriptionRepository interface {
	Create(sub *model.Subscription) error
	FindByEmail(email string) (*model.Subscription, error)
	GetByToken(token string) (*model.Subscription, error)
	Update(sub *model.Subscription) error
	UnsubscribeByToken(token string) error
	GetAllConfirmed() ([]*model.Subscription, error)
}

type Mailer interface {
	SendConfirmation(email, token string) error
	SendWeatherUpdate(email, city string, weatherInfo string) error
}

type SubscriptionService struct {
	Repo   SubscriptionRepository
	Mailer Mailer
}

func NewSubscriptionService(repo SubscriptionRepository, mailer Mailer) *SubscriptionService {
	return &SubscriptionService{Repo: repo, Mailer: mailer}
}

func generateToken() (string, error) {
	return uuid.New().String(), nil
}

func (s *SubscriptionService) Subscribe(sub *model.Subscription) (*model.Subscription, error) {
	existing, err := s.Repo.FindByEmail(sub.Email)
	if err != nil && !errors.Is(err, ErrNotFound) && err.Error() != "record not found" {
		pkg.Logger.Error("failed to check existing subscription", zap.Error(err))
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already subscribed")
	}

	confirmToken, err := generateToken()
	if err != nil {
		pkg.Logger.Error("failed to generate confirm token", zap.Error(err))
		return nil, errors.New("failed generating token")
	}

	unsubscribeToken, err := generateToken()
	if err != nil {
		pkg.Logger.Error("failed to generate unsubscribe token", zap.Error(err))
		return nil, errors.New("failed generating token")
	}

	sub.ConfirmToken = confirmToken
	sub.UnsubscribeToken = unsubscribeToken
	sub.Confirmed = false

	if err := s.Repo.Create(sub); err != nil {
		pkg.Logger.Error("failed to create subscription", zap.Error(err))
		return nil, err
	}

	subCreated, err := s.Repo.FindByEmail(sub.Email)
	if err != nil {
		pkg.Logger.Error("failed to retrieve created subscription", zap.Error(err))
		return nil, err
	}

	if err := s.Mailer.SendConfirmation(subCreated.Email, confirmToken); err != nil {
		pkg.Logger.Error("failed to send confirmation email", zap.String("email", subCreated.Email), zap.Error(err))
	}
	return subCreated, nil
}

func (s *SubscriptionService) ConfirmSubscription(token string) error {
	sub, err := s.Repo.GetByToken(token)
	if err != nil {
		pkg.Logger.Error("failed to get subscription by token", zap.String("token", token), zap.Error(err))
		return errors.New("subscription not found")
	}

	if sub.Confirmed {
		return errors.New("already confirmed")
	}

	sub.Confirmed = true
	if err := s.Repo.Update(sub); err != nil {
		pkg.Logger.Error("failed to update subscription as confirmed", zap.Error(err))
		return err
	}
	return nil
}

func (s *SubscriptionService) Unsubscribe(token string) error {
	if err := s.Repo.UnsubscribeByToken(token); err != nil {
		pkg.Logger.Error("failed to unsubscribe by token", zap.String("token", token), zap.Error(err))
		return err
	}
	return nil
}

func (s *SubscriptionService) GetAllConfirmed() ([]*model.Subscription, error) {
	subs, err := s.Repo.GetAllConfirmed()
	if err != nil {
		pkg.Logger.Error("failed to get all confirmed subscriptions", zap.Error(err))
		return nil, err
	}
	return subs, nil
}

func (s *SubscriptionService) Update(sub *model.Subscription) error {
	if err := s.Repo.Update(sub); err != nil {
		pkg.Logger.Error("failed to update subscription", zap.Error(err))
		return err
	}
	return nil
}

func (s *SubscriptionService) SendWeatherUpdate(email, body string) error {
	if err := s.Mailer.SendWeatherUpdate(email, "", body); err != nil {
		pkg.Logger.Error("failed to send weather update", zap.String("email", email), zap.Error(err))
		return err
	}
	return nil
}
