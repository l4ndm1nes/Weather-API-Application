package port

import (
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
)

type SubscriptionRepository interface {
	Create(sub *model.Subscription) error
	FindByEmail(email string) (*model.Subscription, error)
	GetByToken(token string) (*model.Subscription, error)
	Update(sub *model.Subscription) error
	UnsubscribeByToken(token string) error
}
