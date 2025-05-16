package repo

import (
	"time"
)

type SubscriptionDB struct {
	ID               int64      `gorm:"primaryKey"`
	Email            string     `gorm:"size:255;not null"`
	City             string     `gorm:"size:255;not null"`
	Frequency        string     `gorm:"size:16;not null"`
	Confirmed        bool       `gorm:"not null"`
	ConfirmToken     string     `gorm:"size:255;not null"`
	UnsubscribeToken string     `gorm:"size:255;not null"`
	CreatedAt        time.Time  `gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime"`
	LastSentAt       *time.Time `gorm:"column:last_sent_at"`
}

func (SubscriptionDB) TableName() string {
	return "subscriptions"
}
