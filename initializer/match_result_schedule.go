package initializer

import (
	"context"
	"fmt"
)

type MatchResultScheduleInitializer struct {
	matchService MatchService
	logger       Logger
}

func NewMatchResultScheduleInitializer(matchService MatchService, logger Logger) *MatchResultScheduleInitializer {
	return &MatchResultScheduleInitializer{
		matchService: matchService,
		logger:       logger,
	}
}

func (i *MatchResultScheduleInitializer) ReSchedule(ctx context.Context) error {
	i.logger.Info().Msg("initializing matches to re-schedule")

	matches, err := i.matchService.List(ctx, "scheduled")
	if err != nil {
		return fmt.Errorf("failed to get matches: %w", err)
	}

	i.logger.Info().Msg(fmt.Sprintf("found %d match(es) to re-schedule", len(matches)))

	for j := range matches {
		if errScheduling := i.matchService.ScheduleMatchResultAcquiring(matches[j]); errScheduling != nil {
			i.logger.Error().Err(err).Msg("failed to schedule match for result acquiring")

			if err := i.matchService.Update(ctx, matches[j].ID, "scheduling_error"); err != nil {
				return fmt.Errorf("failed to update match: %w", err)
			}
		}
	}

	return nil
}
