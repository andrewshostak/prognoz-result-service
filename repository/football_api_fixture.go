package repository

import (
	"context"

	"github.com/jackc/pgtype"
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

func (r *FootballAPIFixtureRepository) Update(ctx context.Context, id uint, data pgtype.JSONB) (*FootballApiFixture, error) {
	fixture := FootballApiFixture{ID: id}
	result := r.db.WithContext(ctx).Model(&fixture).Updates(FootballApiFixture{Data: data})
	if result.Error != nil {
		return nil, result.Error
	}

	return &fixture, nil
}
