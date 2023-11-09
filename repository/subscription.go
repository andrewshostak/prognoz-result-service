package repository

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"

	"github.com/andrewshostak/result-service/errs"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(ctx context.Context, subscription Subscription) (*Subscription, error) {
	result := r.db.WithContext(ctx).Create(&subscription)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrForeignKeyViolated) {
			return nil, fmt.Errorf("match id does not exist: %w", errs.WrongMatchIDError{Message: result.Error.Error()})
		}
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return nil, fmt.Errorf("subscription already exists: %w", errs.SubscriptionAlreadyExistsError{Message: result.Error.Error()})
		}

		return nil, result.Error
	}

	return &subscription, nil
}
