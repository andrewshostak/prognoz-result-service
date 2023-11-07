package repository

import (
	"context"
	"fmt"

	"github.com/andrewshostak/result-service/errs"
	"gorm.io/gorm"
)

type AliasRepository struct {
	db *gorm.DB
}

func NewAliasRepository(db *gorm.DB) *AliasRepository {
	return &AliasRepository{db: db}
}

func (r *AliasRepository) Find(ctx context.Context, alias string) (*Alias, error) {
	var a Alias

	result := r.db.WithContext(ctx).Joins("FootballApiTeam").Where("alias ILIKE ?", alias).First(&a)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("alias %s not found: %w", alias, errs.AliasNotFoundError{Message: result.Error.Error()})
		}

		return nil, result.Error
	}

	return &a, nil
}
