package model

import "time"

type Subscription struct {
	ID               int64     `gorm:"primaryKey" json:"id"`
	Email            string    `gorm:"size:255;not null" json:"email"`
	City             string    `gorm:"size:255;not null" json:"city"`
	Frequency        string    `gorm:"size:16;not null" json:"frequency"` // "hourly" or "daily"
	Confirmed        bool      `gorm:"not null" json:"confirmed"`
	ConfirmToken     string    `gorm:"size:255;not null" json:"-"`
	UnsubscribeToken string    `gorm:"size:255;not null" json:"-"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
