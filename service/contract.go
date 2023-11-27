package service

import (
	"context"
	"time"

	"github.com/andrewshostak/result-service/client"
	"github.com/andrewshostak/result-service/repository"
	"github.com/procyon-projects/chrono"
)

type AliasRepository interface {
	Find(ctx context.Context, alias string) (*repository.Alias, error)
}

type MatchRepository interface {
	Create(ctx context.Context, match repository.Match) (*repository.Match, error)
	List(ctx context.Context, resultStatus repository.ResultStatus) ([]repository.Match, error)
	One(ctx context.Context, search repository.Match) (*repository.Match, error)
	Update(ctx context.Context, id uint, resultStatus repository.ResultStatus) (*repository.Match, error)
}

type FootballAPIFixtureRepository interface {
	Create(ctx context.Context, fixture repository.FootballApiFixture, data repository.Data) (*repository.FootballApiFixture, error)
	Update(ctx context.Context, id uint, data repository.Data) (*repository.FootballApiFixture, error)
}

type FootballAPIClient interface {
	SearchFixtures(ctx context.Context, search client.FixtureSearch) (*client.FixturesResponse, error)
}

type SubscriptionRepository interface {
	Create(ctx context.Context, subscription repository.Subscription) (*repository.Subscription, error)
}

type TaskScheduler interface {
	Schedule(task func(ctx context.Context), period time.Duration, startTime time.Time) (*chrono.ScheduledRunnableTask, error)
	Cancel(scheduledTask *chrono.ScheduledRunnableTask)
}
