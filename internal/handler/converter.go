package handler

import (
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
)

func ToDomainFromRequest(req *SubscribeRequest) *model.Subscription {
	if req == nil {
		return nil
	}
	return &model.Subscription{
		Email:     req.Email,
		City:      req.City,
		Frequency: req.Frequency,
	}
}
