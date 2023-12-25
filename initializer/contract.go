package initializer

import (
	"context"

	"github.com/andrewshostak/result-service/service"
	"github.com/rs/zerolog"
)

type MatchService interface {
	List(ctx context.Context, status string) ([]service.Match, error)
	ScheduleMatchResultAcquiring(match service.Match) error
	Update(ctx context.Context, id uint, status string) error
}

type NotifierService interface {
	NotifySubscribers(ctx context.Context) error
}

type Logger interface {
	Error() *zerolog.Event
	Info() *zerolog.Event
}
