package repo

import (
	"github.com/l4ndm1nes/Weather-API-Application/internal/model"
	"github.com/l4ndm1nes/Weather-API-Application/internal/service"
	"github.com/l4ndm1nes/Weather-API-Application/pkg"
	"go.uber.org/zap"
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
		pkg.Logger.Error("Failed to create subscription",
			zap.String("email", sub.Email),
			zap.Error(err),
		)
		return err
	}
	pkg.Logger.Info("Subscription created",
		zap.String("email", sub.Email),
	)
	return nil
}

func (r *PostgresRepo) FindByEmail(email string) (*model.Subscription, error) {
	var dbSub SubscriptionDB
	err := r.db.Where("email = ?", email).First(&dbSub).Error
	if err == gorm.ErrRecordNotFound {
		pkg.Logger.Warn("Subscription not found by email", zap.String("email", email))
		return nil, err
	}
	if err != nil {
		pkg.Logger.Error("Failed to find subscription by email",
			zap.String("email", email),
			zap.Error(err),
		)
		return nil, err
	}
	pkg.Logger.Info("Subscription found by email", zap.String("email", email))
	return ToDomain(&dbSub), nil
}

func (r *PostgresRepo) GetByToken(token string) (*model.Subscription, error) {
	var dbSub SubscriptionDB
	result := r.db.Where("confirm_token = ?", token).First(&dbSub)
	if result.Error == gorm.ErrRecordNotFound {
		pkg.Logger.Warn("Subscription not found by token", zap.String("token", token))
		return nil, result.Error
	}
	if result.Error != nil {
		pkg.Logger.Error("Failed to get subscription by token", zap.String("token", token), zap.Error(result.Error))
		return nil, result.Error
	}
	pkg.Logger.Info("Subscription found by token", zap.String("token", token))
	return ToDomain(&dbSub), nil
}

func (r *PostgresRepo) Update(sub *model.Subscription) error {
	err := r.db.Save(ToDB(sub)).Error
	if err != nil {
		pkg.Logger.Error("Failed to update subscription",
			zap.Int64("id", sub.ID),
			zap.Error(err),
		)
		return err
	}
	pkg.Logger.Info("Subscription updated", zap.Int64("id", sub.ID))
	return nil
}

func (r *PostgresRepo) GetAllConfirmed() ([]*model.Subscription, error) {
	var dbSubs []SubscriptionDB
	err := r.db.Where("confirmed = ?", true).Find(&dbSubs).Error
	if err != nil {
		pkg.Logger.Error("Failed to get all confirmed subscriptions", zap.Error(err))
		return nil, err
	}
	var subs []*model.Subscription
	for _, dbSub := range dbSubs {
		subs = append(subs, ToDomain(&dbSub))
	}
	pkg.Logger.Info("All confirmed subscriptions fetched", zap.Int("count", len(subs)))
	return subs, nil
}

func (r *PostgresRepo) UnsubscribeByToken(token string) error {
	result := r.db.Where("unsubscribe_token = ?", token).Delete(&SubscriptionDB{})
	if result.Error != nil {
		pkg.Logger.Error("Failed to unsubscribe by token",
			zap.String("token", token),
			zap.Error(result.Error),
		)
		return result.Error
	}
	if result.RowsAffected == 0 {
		pkg.Logger.Warn("No subscription found to unsubscribe by token", zap.String("token", token))
		return gorm.ErrRecordNotFound
	}
	pkg.Logger.Info("Unsubscribed by token", zap.String("token", token))
	return nil
}
