package repo

import (
	"github.com/l4ndm1nes/Weather-API-Application/internal/app/port"
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
	"gorm.io/gorm"
)

type PostgresRepo struct {
	db *gorm.DB
}

func NewPostgresRepo(db *gorm.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

var _ port.SubscriptionRepository = (*PostgresRepo)(nil)

func (r *PostgresRepo) Create(sub *model.Subscription) error {
	return r.db.Create(sub).Error
}

func (r *PostgresRepo) FindByEmail(email string) (*model.Subscription, error) {
	var sub model.Subscription
	err := r.db.Where("email = ?", email).First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *PostgresRepo) ConfirmByToken(token string) error {
	return r.db.Model(&model.Subscription{}).
		Where("confirm_token = ?", token).
		Update("confirmed", true).Error
}

func (r *PostgresRepo) UnsubscribeByToken(token string) error {
	result := r.db.Where("unsubscribe_token = ?", token).Delete(&model.Subscription{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *PostgresRepo) GetByToken(token string) (*model.Subscription, error) {
	var sub model.Subscription
	result := r.db.Where("confirm_token = ?", token).First(&sub)
	if result.Error != nil {
		return nil, result.Error
	}
	return &sub, nil
}

func (r *PostgresRepo) Update(sub *model.Subscription) error {
	return r.db.Save(sub).Error
}
