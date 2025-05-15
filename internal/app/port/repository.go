package port

import (
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
)

type SubscriptionRepository interface {
	Create(sub *model.Subscription) error
	FindByEmail(email string) (*model.Subscription, error)
	ConfirmByToken(token string) error
	UnsubscribeByToken(token string) error
}
