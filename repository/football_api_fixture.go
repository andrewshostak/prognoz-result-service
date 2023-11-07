package repository

import (
	"context"

	"gorm.io/gorm"
)

type FootballAPIFixtureRepository struct {
	db *gorm.DB
}

func NewFootballAPIFixtureRepository(db *gorm.DB) *FootballAPIFixtureRepository {
	return &FootballAPIFixtureRepository{db: db}
}

func (r *FootballAPIFixtureRepository) Create(ctx context.Context, fixture FootballApiFixture) (*FootballApiFixture, error) {
	result := r.db.WithContext(ctx).Create(&fixture)
	if result.Error != nil {
		return nil, result.Error
	}

	return &fixture, nil
}
