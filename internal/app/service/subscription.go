package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/l4ndm1nes/Weather-API-Application/internal/app/port"
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
)

type SubscriptionService struct {
	Repo port.SubscriptionRepository
}

func NewSubscriptionService(repo port.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{Repo: repo}
}

func generateToken() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *SubscriptionService) Subscribe(email, city, frequency string) (*model.Subscription, error) {
	existing, _ := s.Repo.FindByEmail(email)
	if existing != nil {
		return nil, errors.New("email already subscribed")
	}

	confirmToken, _ := generateToken()
	unsubscribeToken, _ := generateToken()

	sub := &model.Subscription{
		Email:            email,
		City:             city,
		Frequency:        frequency,
		Confirmed:        false,
		ConfirmToken:     confirmToken,
		UnsubscribeToken: unsubscribeToken,
	}

	if err := s.Repo.Create(sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *SubscriptionService) ConfirmSubscription(token string) error {
	return s.Repo.ConfirmByToken(token)
}

func (s *SubscriptionService) Unsubscribe(token string) error {
	return s.Repo.UnsubscribeByToken(token)
}
