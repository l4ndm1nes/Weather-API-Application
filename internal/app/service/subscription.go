package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/l4ndm1nes/Weather-API-Application/internal/app/port"
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
)

var ErrNotFound = errors.New("subscription not found")

type SubscriptionService struct {
	Repo   port.SubscriptionRepository
	Mailer port.Mailer
}

func NewSubscriptionService(repo port.SubscriptionRepository, mailer port.Mailer) *SubscriptionService {
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

	_ = s.Mailer.SendConfirmation(email, confirmToken)
	return sub, nil
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
