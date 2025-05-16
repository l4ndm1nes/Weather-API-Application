package repo

import (
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
	"github.com/l4ndm1nes/Weather-API-Application/internal/service"
	"gorm.io/gorm"
)

type PostgresRepo struct {
	db *gorm.DB
}

func NewPostgresRepo(db *gorm.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

var _ service.SubscriptionRepository = (*PostgresRepo)(nil)

func (r *PostgresRepo) Create(sub *model.Subscription) error {
	dbSub := ToDB(sub)
	err := r.db.Create(dbSub).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *PostgresRepo) FindByEmail(email string) (*model.Subscription, error) {
	var dbSub SubscriptionDB
	err := r.db.Where("email = ?", email).First(&dbSub).Error
	if err != nil {
		return nil, err
	}
	return ToDomain(&dbSub), nil
}

func (r *PostgresRepo) GetByToken(token string) (*model.Subscription, error) {
	var dbSub SubscriptionDB
	result := r.db.Where("confirm_token = ?", token).First(&dbSub)
	if result.Error != nil {
		return nil, result.Error
	}
	return ToDomain(&dbSub), nil
}

func (r *PostgresRepo) Update(sub *model.Subscription) error {
	return r.db.Save(ToDB(sub)).Error
}

func (r *PostgresRepo) GetAllConfirmed() ([]*model.Subscription, error) {
	var dbSubs []SubscriptionDB
	if err := r.db.Where("confirmed = ?", true).Find(&dbSubs).Error; err != nil {
		return nil, err
	}
	var subs []*model.Subscription
	for _, dbSub := range dbSubs {
		subs = append(subs, ToDomain(&dbSub))
	}
	return subs, nil
}

func (r *PostgresRepo) UnsubscribeByToken(token string) error {
	result := r.db.Where("unsubscribe_token = ?", token).Delete(&SubscriptionDB{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
