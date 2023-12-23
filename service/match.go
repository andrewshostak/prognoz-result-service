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
const stateMatchFinished = "Match Finished"

type MatchService struct {
	aliasRepository              AliasRepository
	matchRepository              MatchRepository
	footballAPIFixtureRepository FootballAPIFixtureRepository
	footballAPIClient            FootballAPIClient
	taskScheduler                TaskScheduler
	pollingMaxRetries            uint
	pollingInterval              time.Duration
	pollingFirstAttemptDelay     time.Duration
}

func NewMatchService(
	aliasRepository AliasRepository,
	matchRepository MatchRepository,
	footballAPIFixtureRepository FootballAPIFixtureRepository,
	footballAPIClient FootballAPIClient,
	taskScheduler TaskScheduler,
	pollingMaxRetries uint,
	pollingInterval time.Duration,
	pollingFirstAttemptDelay time.Duration,
) *MatchService {
	return &MatchService{
		aliasRepository:              aliasRepository,
		matchRepository:              matchRepository,
		footballAPIFixtureRepository: footballAPIFixtureRepository,
		footballAPIClient:            footballAPIClient,
		taskScheduler:                taskScheduler,
		pollingMaxRetries:            pollingMaxRetries,
		pollingInterval:              pollingInterval,
		pollingFirstAttemptDelay:     pollingFirstAttemptDelay,
	}
}

func (s *MatchService) Create(ctx context.Context, request CreateMatchRequest) (uint, error) {
	aliasHome, err := s.findAlias(ctx, request.AliasHome)
	if err != nil {
		return 0, fmt.Errorf("failed to find home team alias: %w", err)
	}

	aliasAway, err := s.findAlias(ctx, request.AliasAway)
	if err != nil {
		return 0, fmt.Errorf("failed to find away team alias: %w", err)
	}

	match, err := s.matchRepository.One(ctx, repository.Match{
		StartsAt:   request.StartsAt.UTC(),
		HomeTeamID: aliasHome.TeamID,
		AwayTeamID: aliasAway.TeamID,
	})

	// if match already exists, we can return id immediately and don't call external api
	if match != nil {
		return match.ID, nil
	}

	if !errors.As(err, &errs.MatchNotFoundError{}) {
		return 0, fmt.Errorf("unexpected error when getting a match: %w", err)
	}

	fmt.Printf("match between %s and %s is not found in the database. making an attempt to find it in external api. \n", request.AliasHome, request.AliasAway)

	date := request.StartsAt.UTC().Format(dateFormat)
	season := uint(s.getSeason())
	response, err := s.footballAPIClient.SearchFixtures(ctx, client.FixtureSearch{
		Season:   season,
		Timezone: time.UTC.String(),
		Date:     &date,
		TeamID:   &aliasHome.FootballApiTeam.ID,
	})
	if err != nil {
		return 0, fmt.Errorf("unable to search fixtures in external api: %w", err)
	}

	if len(response.Response) < 1 {
		return 0, errs.UnexpectedNumberOfItemsError{Message: fmt.Sprintf("fixture starting at %s with team id %d is not found in external api", date, aliasHome.FootballApiTeam.ID)}
	}

	fixture := fromClientFootballAPIFixture(response.Response[0])

	if fixture.Fixture.Status.Long == stateMatchFinished {
		return 0, fmt.Errorf("%s: %w", fmt.Sprintf("status of the fixture with external id %d is %s", fixture.Fixture.ID, stateMatchFinished), errs.ErrIncorrectFixtureStatus)
	}

	startsAt, err := time.Parse(time.RFC3339, fixture.Fixture.Date)
	if err != nil {
		return 0, fmt.Errorf("unable to parse received from external api fixture date %s: %w", fixture.Fixture.Date, err)
	}

	toCreate := repository.Match{HomeTeamID: aliasHome.TeamID, AwayTeamID: aliasAway.TeamID, StartsAt: startsAt}
	created, err := s.matchRepository.Create(ctx, toCreate)
	if err != nil {
		return 0, fmt.Errorf("failed to create match with team ids %d and %d starting at %s: %w", aliasHome.TeamID, aliasAway.TeamID, startsAt, err)
	}

	createdFixture, err := s.footballAPIFixtureRepository.Create(ctx, repository.FootballApiFixture{
		ID:      fixture.Fixture.ID,
		MatchID: created.ID,
	}, toRepositoryFootballAPIFixtureData(fixture))
	if err != nil {
		return 0, fmt.Errorf("failed to create football api fixture with match id %d: %w", created.ID, err)
	}

	mappedMatch, err := fromRepositoryMatch(*created)
	if err != nil {
		return 0, fmt.Errorf("failed to map from repository match: %w", err)
	}

	mappedFixture, err := fromRepositoryFootballAPIFixture(*createdFixture)
	if err != nil {
		return 0, fmt.Errorf("failed to map from repository api fixture: %w", err)
	}

	if err := s.scheduleMatchResultAcquiring(matchResultTaskParams{
		match:     *mappedMatch,
		fixture:   *mappedFixture,
		aliasHome: *aliasHome,
		aliasAway: *aliasAway,
		season:    season,
	}); err != nil {
		return 0, fmt.Errorf("failed to schedule match result aquiring: %w", err)
	}

	_, err = s.matchRepository.Update(ctx, created.ID, repository.Scheduled)
	if err != nil {
		return 0, fmt.Errorf("failed to set match status to %s: %w", repository.Scheduled, err)
	}

	return created.ID, nil
}

func (s *MatchService) List(ctx context.Context, status string) ([]Match, error) {
	resultStatus := repository.ResultStatus(status)
	matches, err := s.matchRepository.List(ctx, resultStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to list matches with %s result status: %w", resultStatus, err)
	}

	mapped, err := fromRepositoryMatches(matches)
	if err != nil {
		return nil, fmt.Errorf("failed to map from repository matches: %w", err)
	}

	return mapped, nil
}

func (s *MatchService) ScheduleMatchResultAcquiring(match Match) error {
	if len(match.FootballApiFixtures) < 1 {
		return errors.New("match relation football api fixtures are not found")
	}

	if match.HomeTeam == nil || match.AwayTeam == nil {
		return errors.New("match relations home/away team are not found")
	}

	if len(match.HomeTeam.Aliases) < 1 {
		return errors.New("match relation home team doesn't have aliases")
	}

	if len(match.AwayTeam.Aliases) < 1 {
		return errors.New("match relation away team doesn't have aliases")
	}

	params := matchResultTaskParams{
		match:     Match{ID: match.ID, StartsAt: match.StartsAt},
		fixture:   match.FootballApiFixtures[0],
		aliasHome: match.HomeTeam.Aliases[0],
		aliasAway: match.AwayTeam.Aliases[0],
		season:    uint(s.getSeason()),
	}
	return s.scheduleMatchResultAcquiring(params)
}

func (s *MatchService) Update(ctx context.Context, id uint, status string) error {
	resultStatus := repository.ResultStatus(status)
	_, err := s.matchRepository.Update(ctx, id, resultStatus)
	if err != nil {
		return fmt.Errorf("failed to set match status to %s: %w", repository.Scheduled, err)
	}

	return nil
}

func (s *MatchService) findAlias(ctx context.Context, alias string) (*Alias, error) {
	foundAlias, err := s.aliasRepository.Find(ctx, alias)
	if err != nil {
		return nil, fmt.Errorf("failed to find team alias: %w", err)
	}

	if foundAlias.FootballApiTeam == nil {
		return nil, errors.New(fmt.Sprintf("alias %s found, but there is no releated external(football api) team", alias))
	}

	mapped := fromRepositoryAlias(*foundAlias)
	return &mapped, nil
}

func (s *MatchService) scheduleMatchResultAcquiring(params matchResultTaskParams) error {
	matchDetails := fmt.Sprintf(
		"match with id %d between %s and %s starting at %s",
		params.match.ID,
		params.aliasHome.Alias,
		params.aliasAway.Alias,
		params.match.StartsAt,
	)

	fmt.Printf("scheduling a task for %s \n", matchDetails)

	i := 1
	ch := make(chan resultTaskChan)
	search := client.FixtureSearch{Season: params.season, Timezone: time.UTC.String(), ID: &params.fixture.ID}

	key := getTaskKey(params.match.ID, params.fixture.ID)
	err := s.taskScheduler.Schedule(key, s.getTaskFunc(i, ch, search, matchDetails), s.pollingInterval, params.match.StartsAt.Add(s.pollingFirstAttemptDelay))
	if err != nil {
		return fmt.Errorf("failed to schedule a task for %s: %w", matchDetails, err)
	}

	go s.handleTaskResult(context.Background(), ch, params.fixture.ID, params.match.ID, matchDetails)

	return nil
}

// getSeason returns current year if current time is after June 1, otherwise previous year
func (s *MatchService) getSeason() int {
	now := time.Now()

	seasonBound := time.Date(now.Year(), 6, 1, 0, 0, 0, 0, time.UTC)

	if now.After(seasonBound) {
		return now.Year()
	}

	return now.AddDate(-1, 0, 0).Year()
}

func (s *MatchService) getTaskFunc(i int, ch chan<- resultTaskChan, search client.FixtureSearch, matchDetails string) func(c context.Context) {
	return func(c context.Context) {
		fmt.Printf("iteration %d for %s \n", i, matchDetails)

		response, err := s.footballAPIClient.SearchFixtures(c, search)
		if err != nil {
			fmt.Printf("received error when searching fixtures for match %s. cancelling. error: %s \n", matchDetails, err.Error())
			i++

			if s.retriesLimitReached(i) {
				s.writeError(matchDetails, ch)
			}
			return
		}

		if len(response.Response) < 1 {
			fmt.Printf("unexpected length of fixture search result for match %s \n", matchDetails)
			i++

			if s.retriesLimitReached(i) {
				s.writeError(matchDetails, ch)
			}
			return
		}

		fixture := fromClientFootballAPIFixture(response.Response[0])

		if fixture.Fixture.Status.Long != stateMatchFinished {
			fmt.Printf("status is not finished (got \"%s\") for %s\n", fixture.Fixture.Status.Long, matchDetails)
			i++

			if s.retriesLimitReached(i) {
				s.writeError(matchDetails, ch)
			}
			return
		}

		fmt.Printf("received result %d:%d for the match %s. cancelling the task \n", fixture.Goals.Home, fixture.Goals.Away, matchDetails)
		ch <- resultTaskChan{fixture: &fixture}
		close(ch)
	}
}

func (s *MatchService) handleTaskResult(
	ctx context.Context,
	ch <-chan resultTaskChan,
	fixtureID uint,
	matchID uint,
	matchDetails string,
) {
	result := <-ch
	key := getTaskKey(matchID, fixtureID)
	s.taskScheduler.Cancel(key)

	fmt.Printf("scheduled task cancelled for %s \n", matchDetails)

	if result.error != nil {
		_, err := s.matchRepository.Update(ctx, matchID, repository.Error)
		if err != nil {
			fmt.Printf("failed to update result status to %s for %s: %s \n", repository.Error, matchDetails, err.Error())
			return
		}

		return
	}

	if _, err := s.footballAPIFixtureRepository.Update(ctx, fixtureID, toRepositoryFootballAPIFixtureData(*result.fixture)); err != nil {
		fmt.Printf("failed to update fixture for %s: %s \n", matchDetails, err.Error())
		return
	}

	if _, err := s.matchRepository.Update(ctx, matchID, repository.Successful); err != nil {
		fmt.Printf("failed to update result status to %s for %s: %s \n", repository.Successful, matchDetails, err.Error())
		return
	}

	return
}

func (s *MatchService) retriesLimitReached(i int) bool {
	return i > int(s.pollingMaxRetries)
}

func (s *MatchService) writeError(matchDetails string, ch chan<- resultTaskChan) {
	errMessage := fmt.Sprintf("retries limit (%d) reached for %s. cancelling", s.pollingMaxRetries, matchDetails)
	fmt.Println(errMessage)
	ch <- resultTaskChan{error: errors.New(errMessage)}
}

func getTaskKey(matchID uint, fixtureID uint) string {
	return fmt.Sprintf("%d-%d", matchID, fixtureID)
}

type matchResultTaskParams struct {
	match     Match
	fixture   FootballAPIFixture
	aliasHome Alias
	aliasAway Alias
	season    uint
}

type resultTaskChan struct {
	fixture *Data
	error   error
}
