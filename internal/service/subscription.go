package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
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
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *SubscriptionService) Subscribe(sub *model.Subscription) (*model.Subscription, error) {
	existing, _ := s.Repo.FindByEmail(sub.Email)
	if existing != nil {
		return nil, errors.New("email already subscribed")
	}

	confirmToken, err := generateToken()
	if err != nil {
		return nil, errors.New("failed generating token")
	}

	unsubscribeToken, err := generateToken()
	if err != nil {
		return nil, errors.New("failed generating token")
	}

	sub.ConfirmToken = confirmToken
	sub.UnsubscribeToken = unsubscribeToken
	sub.Confirmed = false

	if err := s.Repo.Create(sub); err != nil {
		return nil, err
	}

	subCreated, err := s.Repo.FindByEmail(sub.Email)
	if err != nil {
		return nil, err
	}

	_ = s.Mailer.SendConfirmation(subCreated.Email, confirmToken)
	return subCreated, nil
}

func (s *SubscriptionService) ConfirmSubscription(token string) error {
	sub, err := s.Repo.GetByToken(token)
	if err != nil {
		return ErrNotFound
	}
	if sub.Confirmed {
		return errors.New("already confirmed")
	}
	sub.Confirmed = true
	return s.Repo.Update(sub)
}

func (s *SubscriptionService) Unsubscribe(token string) error {
	return s.Repo.UnsubscribeByToken(token)
}

func (s *SubscriptionService) GetAllConfirmed() ([]*model.Subscription, error) {
	return s.Repo.GetAllConfirmed()
}

func (s *SubscriptionService) Update(sub *model.Subscription) error {
	return s.Repo.Update(sub)
}

func (s *SubscriptionService) SendWeatherUpdate(email, body string) error {
	return s.Mailer.SendWeatherUpdate(email, "", body)
}
