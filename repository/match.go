package repository

import (
	"context"
	"fmt"

	"github.com/andrewshostak/result-service/errs"
	"gorm.io/gorm"
)

type MatchRepository struct {
	db *gorm.DB
}

func NewMatchRepository(db *gorm.DB) *MatchRepository {
	return &MatchRepository{db: db}
}

func (r *MatchRepository) Create(ctx context.Context, match Match) (*Match, error) {
	result := r.db.WithContext(ctx).Create(&match)
	if result.Error != nil {
		return nil, result.Error
	}

	return &match, nil
}

func (r *MatchRepository) Search(ctx context.Context, search Match) (*Match, error) {
	var match Match

	result := r.db.WithContext(ctx).
		Where(&Match{StartsAt: search.StartsAt, HomeTeamID: search.HomeTeamID, AwayTeamID: search.AwayTeamID}).
		First(&match)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			message := fmt.Sprintf("match between teams with ids %d and %d starting at %s is not found", search.HomeTeamID, search.AwayTeamID, search.StartsAt)
			return nil, fmt.Errorf("%s: %w", message, errs.MatchNotFoundError{Message: result.Error.Error()})
		}

		return nil, result.Error
	}

	return &match, nil
}
