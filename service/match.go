package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/andrewshostak/result-service/client"
	"github.com/andrewshostak/result-service/errs"
	"github.com/andrewshostak/result-service/repository"
)

const dateFormat = "2006-01-02"
const stateNotStarted = "NS"

type MatchService struct {
	aliasRepository   AliasRepository
	matchRepository   MatchRepository
	footballAPIClient FootballAPIClient
	location          *time.Location
}

func NewMatchService(
	aliasRepository AliasRepository,
	matchRepository MatchRepository,
	footballAPIClient FootballAPIClient,
	location *time.Location,
) *MatchService {
	return &MatchService{
		aliasRepository:   aliasRepository,
		matchRepository:   matchRepository,
		footballAPIClient: footballAPIClient,
		location:          location,
	}
}

func (s *MatchService) Create(ctx context.Context, request CreateMatchRequest) (uint, error) {
	aliasHome, err := s.aliasRepository.Find(ctx, request.AliasHome)
	if err != nil {
		return 0, fmt.Errorf("failed to find home team alias: %w", err)
	}

	aliasAway, err := s.aliasRepository.Find(ctx, request.AliasAway)
	if err != nil {
		return 0, fmt.Errorf("failed to find away team alias: %w", err)
	}

	match, err := s.matchRepository.Search(ctx, repository.Match{
		StartsAt:   request.StartsAt,
		HomeTeamID: aliasHome.TeamID,
		AwayTeamID: aliasAway.TeamID,
	})

	if match != nil {
		return match.ID, nil
	}

	if !errors.As(err, &errs.MatchNotFoundError{}) {
		return 0, fmt.Errorf("unexpected error when searching a match: %w", err)
	}

	fmt.Printf("match between %s and %s is not found in the database. making an attempt to find it in the external api. \n", request.AliasHome, request.AliasAway)

	date := request.StartsAt.Format(dateFormat)
	response, err := s.footballAPIClient.SearchFixtures(ctx, client.FixtureSearch{
		Season:   uint(s.getSeason()),
		Timezone: s.location.String(),
		Date:     &date,
		TeamID:   &aliasHome.FootballApiTeam.ID,
	})

	if err != nil {
		return 0, fmt.Errorf("unable to search fixtures: %w", err)
	}

	if len(response.Response) < 1 {
		return 0, errs.UnexpectedNumberOfItemsError{Message: fmt.Sprintf("fixture starting at %s with team id %d is not found", date, aliasHome.FootballApiTeam.ID)}
	}

	if response.Response[0].Fixture.Status.Short != stateNotStarted {
		return 0, fmt.Errorf("%s: %w", fmt.Sprintf("status of the fixture with id %d is not %s", response.Response[0].Fixture.ID, stateNotStarted), errs.ErrIncorrectFixtureStatus)
	}

	return response.Response[0].Fixture.ID, nil
}

// getSeason returns current year if current time is after June 1, otherwise previous year
func (s *MatchService) getSeason() int {
	now := time.Now().In(s.location)

	seasonBound := time.Date(now.Year(), 6, 1, 0, 0, 0, 0, s.location)

	if now.After(seasonBound) {
		return now.Year()
	}

	return now.AddDate(-1, 0, 0).Year()
}
