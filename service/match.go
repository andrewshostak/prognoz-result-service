package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/andrewshostak/result-service/client"
	"github.com/andrewshostak/result-service/errs"
	"github.com/andrewshostak/result-service/repository"
	"github.com/jackc/pgtype"
	"github.com/procyon-projects/chrono"
)

const dateFormat = "2006-01-02"
const (
	stateNotStarted    = "NS"
	stateMatchFinished = "Match Finished"
)
const numberOfRetries = 5
const timeBetweenRetries = 15 * time.Minute
const firstAttemptDelay = 100 * time.Minute

type MatchService struct {
	aliasRepository              AliasRepository
	matchRepository              MatchRepository
	footballAPIFixtureRepository FootballAPIFixtureRepository
	footballAPIClient            FootballAPIClient
	taskScheduler                TaskScheduler
	location                     *time.Location
}

func NewMatchService(
	aliasRepository AliasRepository,
	matchRepository MatchRepository,
	footballAPIFixtureRepository FootballAPIFixtureRepository,
	footballAPIClient FootballAPIClient,
	taskScheduler TaskScheduler,
	location *time.Location,
) *MatchService {
	return &MatchService{
		aliasRepository:              aliasRepository,
		matchRepository:              matchRepository,
		footballAPIFixtureRepository: footballAPIFixtureRepository,
		footballAPIClient:            footballAPIClient,
		taskScheduler:                taskScheduler,
		location:                     location,
	}
}

func (s *MatchService) Create(ctx context.Context, request CreateMatchRequest) (uint, error) {
	aliasHome, err := s.aliasRepository.Find(ctx, request.AliasHome)
	if err != nil {
		return 0, fmt.Errorf("failed to find home team alias: %w", err)
	}

	if aliasHome.FootballApiTeam == nil {
		return 0, errors.New(fmt.Sprintf("alias %s found, but there is no releated external(football api) team", aliasHome.Alias))
	}

	aliasAway, err := s.aliasRepository.Find(ctx, request.AliasAway)
	if err != nil {
		return 0, fmt.Errorf("failed to find away team alias: %w", err)
	}

	if aliasAway.FootballApiTeam == nil {
		return 0, errors.New(fmt.Sprintf("alias %s found, but there is no releated external(football api) team", aliasAway.Alias))
	}

	match, err := s.matchRepository.One(ctx, repository.Match{
		StartsAt:   request.StartsAt,
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

	date := request.StartsAt.Format(dateFormat)
	season := uint(s.getSeason())
	timezone := s.location.String()
	response, err := s.footballAPIClient.SearchFixtures(ctx, client.FixtureSearch{
		Season:   season,
		Timezone: timezone,
		Date:     &date,
		TeamID:   &aliasHome.FootballApiTeam.ID,
	})

	if err != nil {
		return 0, fmt.Errorf("unable to search fixtures in external api: %w", err)
	}

	if len(response.Response) < 1 {
		return 0, errs.UnexpectedNumberOfItemsError{Message: fmt.Sprintf("fixture starting at %s with team id %d is not found in external api", date, aliasHome.FootballApiTeam.ID)}
	}

	fixture := response.Response[0]

	if fixture.Fixture.Status.Short != stateNotStarted {
		return 0, fmt.Errorf("%s: %w", fmt.Sprintf("status of the fixture with external id %d is not %s", fixture.Fixture.ID, stateNotStarted), errs.ErrIncorrectFixtureStatus)
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

	fixtureAsJson, err := toJsonB(fixture)
	if err != nil {
		return 0, err
	}

	createdFixture, err := s.footballAPIFixtureRepository.Create(ctx, repository.FootballApiFixture{
		ID:      fixture.Fixture.ID,
		MatchID: created.ID,
		Data:    *fixtureAsJson,
	})

	if err != nil {
		return 0, fmt.Errorf("failed to create football api fixture with match id %d: %w", created.ID, err)
	}

	createdFixture.Match = created
	if err := s.scheduleMatchResultAcquiring(matchResultTaskParams{
		fixture:   fromRepositoryFootballAPIFixture(*createdFixture),
		aliasHome: fromRepositoryAlias(*aliasHome),
		aliasAway: fromRepositoryAlias(*aliasAway),
		season:    season,
		timezone:  timezone,
	}); err != nil {
		return 0, fmt.Errorf("failed to schedule match result aquiring: %w", err)
	}

	_, err = s.matchRepository.Update(ctx, created.ID, repository.Scheduled)
	if err != nil {
		return 0, fmt.Errorf("failed to set match status to %s: %w", repository.Scheduled, err)
	}

	return created.ID, nil
}

func (s *MatchService) scheduleMatchResultAcquiring(params matchResultTaskParams) error {
	matchDetails := fmt.Sprintf("match with id %d between %s and %s starting at %s", params.fixture.Match.ID, params.aliasHome.Alias, params.aliasAway.Alias, params.fixture.Match.StartsAt)

	fmt.Printf("scheduling a task for %s \n", matchDetails)

	i := 1
	ch := make(chan resultTaskChan)
	search := client.FixtureSearch{Season: params.season, Timezone: params.timezone, ID: &params.fixture.ID}

	scheduledTask, err := s.taskScheduler.Schedule(s.getTaskFunc(i, ch, search, matchDetails), timeBetweenRetries, params.fixture.Match.StartsAt.Add(firstAttemptDelay))
	if err != nil {
		return fmt.Errorf("failed to schedule a task for %s: %w", matchDetails, err)
	}

	go s.handleTaskResult(context.Background(), scheduledTask, ch, params.fixture, matchDetails)

	return nil
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

func (s *MatchService) getTaskFunc(i int, ch chan<- resultTaskChan, search client.FixtureSearch, matchDetails string) func(c context.Context) {
	return func(c context.Context) {
		fmt.Printf("iteration %d for %s \n", i, matchDetails)

		response, err := s.footballAPIClient.SearchFixtures(c, search)
		if err != nil {
			fmt.Printf("received error when searching fixtures for match %s. cancelling. error: %s \n", matchDetails, err.Error())
			i++

			if retriesLimitReached(i) {
				writeError(matchDetails, ch)
			}
			return
		}

		if len(response.Response) < 1 {
			fmt.Printf("unexpected length of fixture search result for match %s \n", matchDetails)
			i++

			if retriesLimitReached(i) {
				writeError(matchDetails, ch)
			}
			return
		}

		if response.Response[0].Fixture.Status.Long != stateMatchFinished {
			fmt.Printf("status is not finished (got \"%s\") for %s\n", response.Response[0].Fixture.Status.Long, matchDetails)
			i++

			if retriesLimitReached(i) {
				writeError(matchDetails, ch)
			}
			return
		}

		fmt.Printf("received result %d:%d for the match %s. cancelling the task \n", response.Response[0].Score.Fulltime.Home, response.Response[0].Score.Fulltime.Away, matchDetails)
		f := fromClientFootballAPIFixture(response.Response[0])
		ch <- resultTaskChan{fixture: &f}
		close(ch)
	}
}

func (s *MatchService) handleTaskResult(
	ctx context.Context,
	scheduledTask *chrono.ScheduledRunnableTask,
	ch <-chan resultTaskChan,
	fixture FootballAPIFixture,
	matchDetails string,
) {
	result := <-ch
	s.taskScheduler.Cancel(scheduledTask)

	fmt.Printf("scheduled task cancelled for %s \n", matchDetails)

	if result.error != nil {
		_, err := s.matchRepository.Update(ctx, fixture.Match.ID, repository.Error)
		if err != nil {
			fmt.Printf("failed to update result status to %s for %s: %s \n", repository.Error, matchDetails, err.Error())
			return
		}

		return
	}

	fixtureAsJson, err := toJsonB(result.fixture)
	if err != nil {
		fmt.Printf("failed to set fixture data as json for %s: %s \n", matchDetails, err.Error())
		return
	}

	_, err = s.footballAPIFixtureRepository.Update(ctx, fixture.ID, *fixtureAsJson)
	if err != nil {
		fmt.Printf("failed to update fixture for %s: %s \n", matchDetails, err.Error())
		return
	}

	_, err = s.matchRepository.Update(ctx, fixture.Match.ID, repository.Successful)
	if err != nil {
		fmt.Printf("failed to update result status to %s for %s: %s \n", repository.Successful, matchDetails, err.Error())
		return
	}

	return
}

func retriesLimitReached(i int) bool {
	return i > numberOfRetries
}

func toJsonB(result interface{}) (*pgtype.JSONB, error) {
	var fixtureAsJson pgtype.JSONB
	if err := fixtureAsJson.Set(result); err != nil {
		return nil, err
	}

	return &fixtureAsJson, nil
}

func writeError(matchDetails string, ch chan<- resultTaskChan) {
	errMessage := fmt.Sprintf("retries limit (%d) reached for %s. cancelling", numberOfRetries, matchDetails)
	fmt.Println(errMessage)
	ch <- resultTaskChan{error: errors.New(errMessage)}
}

type matchResultTaskParams struct {
	fixture   FootballAPIFixture
	aliasHome Alias
	aliasAway Alias
	season    uint
	timezone  string
}

type resultTaskChan struct {
	fixture *Result
	error   error
}
