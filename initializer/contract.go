package initializer

import (
	"context"

	"github.com/andrewshostak/result-service/service"
)

type MatchService interface {
	List(ctx context.Context, status string) ([]service.Match, error)
	ScheduleMatchResultAcquiring(match service.Match) error
	Update(ctx context.Context, id uint, status string) error
}
