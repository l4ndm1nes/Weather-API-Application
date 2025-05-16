package repo

import (
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
)

func ToDomain(subDB *SubscriptionDB) *model.Subscription {
	return &model.Subscription{
		ID:               subDB.ID,
		Email:            subDB.Email,
		City:             subDB.City,
		Frequency:        subDB.Frequency,
		Confirmed:        subDB.Confirmed,
		ConfirmToken:     subDB.ConfirmToken,
		UnsubscribeToken: subDB.UnsubscribeToken,
		CreatedAt:        subDB.CreatedAt,
		UpdatedAt:        subDB.UpdatedAt,
		LastSentAt:       subDB.LastSentAt,
	}
}

func ToDB(sub *model.Subscription) *SubscriptionDB {
	return &SubscriptionDB{
		ID:               sub.ID,
		Email:            sub.Email,
		City:             sub.City,
		Frequency:        sub.Frequency,
		Confirmed:        sub.Confirmed,
		ConfirmToken:     sub.ConfirmToken,
		UnsubscribeToken: sub.UnsubscribeToken,
		CreatedAt:        sub.CreatedAt,
		UpdatedAt:        sub.UpdatedAt,
		LastSentAt:       sub.LastSentAt,
	}
}
