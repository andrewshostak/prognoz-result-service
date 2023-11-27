package initializer

import (
	"context"
	"fmt"
)

type MatchResultScheduleInitializer struct {
	matchService MatchService
}

func NewMatchResultScheduleInitializer(matchService MatchService) *MatchResultScheduleInitializer {
	return &MatchResultScheduleInitializer{
		matchService: matchService,
	}
}

func (i *MatchResultScheduleInitializer) ReSchedule(ctx context.Context) error {
	fmt.Printf("initializing matches to re-schedule \n")

	matches, err := i.matchService.List(ctx, "scheduled")
	if err != nil {
		return fmt.Errorf("failed to get matches: %w", err)
	}

	fmt.Printf("number of found matches to re-schedule: %d \n", len(matches))

	for j := range matches {
		if err := i.matchService.ScheduleMatchResultAcquiring(matches[j]); err != nil {
			fmt.Printf("failed to schedule match for result acquiring: %s \n", err.Error())
		}

		if err := i.matchService.Update(ctx, matches[j].ID, "scheduling_error"); err != nil {
			return fmt.Errorf("failed to update match: %w", err)
		}
	}

	return nil
}
