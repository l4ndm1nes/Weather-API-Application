package model

import "time"

type Subscription struct {
	ID               int64
	Email            string
	City             string
	Frequency        string
	Confirmed        bool
	ConfirmToken     string
	UnsubscribeToken string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	LastSentAt       *time.Time
}
