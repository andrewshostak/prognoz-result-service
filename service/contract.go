package service

import (
	"context"

	"github.com/andrewshostak/result-service/client"
	"github.com/andrewshostak/result-service/repository"
)

type AliasRepository interface {
	Find(ctx context.Context, alias string) (*repository.Alias, error)
}

type MatchRepository interface {
	Create(ctx context.Context, match repository.Match) (*repository.Match, error)
	One(ctx context.Context, search repository.Match) (*repository.Match, error)
}

type FootballAPIFixtureRepository interface {
	Create(ctx context.Context, fixture repository.FootballApiFixture) (*repository.FootballApiFixture, error)
}

type FootballAPIClient interface {
	SearchFixtures(ctx context.Context, search client.FixtureSearch) (*client.FixturesResponse, error)
}
