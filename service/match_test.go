package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/andrewshostak/result-service/repository"
	"github.com/andrewshostak/result-service/service"
	"github.com/andrewshostak/result-service/service/mocks"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestMatchService_List(t *testing.T) {
	aliasRepository := mocks.NewAliasRepository(t)
	matchRepository := mocks.NewMatchRepository(t)
	footballAPIFixtureRepository := mocks.NewFootballAPIFixtureRepository(t)
	footballAPIClient := mocks.NewFootballAPIClient(t)
	taskScheduler := mocks.NewTaskScheduler(t)
	logger := mocks.NewLogger(t)

	pollingMaxRetries := uint(5)
	pollingInterval := 15 * time.Minute
	pollingFirstAttemptDelay := 115 * time.Minute

	ms := service.NewMatchService(
		aliasRepository,
		matchRepository,
		footballAPIFixtureRepository,
		footballAPIClient,
		taskScheduler,
		logger,
		pollingMaxRetries,
		pollingInterval,
		pollingFirstAttemptDelay,
	)

	ctx := context.Background()
	status := "scheduled"

	t.Run("it should return wrapped error if list method returns error", func(t *testing.T) {
		errRepo := errors.New(gofakeit.Sentence(2))
		matchRepository.On("List", ctx, repository.ResultStatus(status)).Return(nil, errRepo).Once()

		result, err := ms.List(ctx, status)
		assert.EqualError(t, err, fmt.Sprintf("failed to list matches with %s result status: %s", status, errRepo.Error()))
		assert.Nil(t, result)
	})

	t.Run("it should return mapped matches if list method returns matches", func(t *testing.T) {
		repoList := []repository.Match{fakeRepositoryMatch(true, true), fakeRepositoryMatch(true, true)}
		matchRepository.On("List", ctx, repository.ResultStatus(status)).Return(repoList, nil).Once()

		result, err := ms.List(ctx, status)
		assert.NoError(t, err)
		assert.Equal(t, []service.Match{expectedMatch(repoList[0]), expectedMatch(repoList[1])}, result)
	})
}

func fakeRepositoryMatch(teams bool, fixtures bool) repository.Match {
	matchID := uint(gofakeit.Uint8())
	homeTeamID := uint(gofakeit.Uint8())
	awayTeamID := uint(gofakeit.Uint8())

	var homeTeam *repository.Team
	var awayTeam *repository.Team
	var apiFixtures []repository.FootballApiFixture

	if teams {
		fakeTeam1, fakeTeam2 := fakeRepositoryTeam(homeTeamID, true), fakeRepositoryTeam(awayTeamID, true)
		homeTeam = &fakeTeam1
		awayTeam = &fakeTeam2
	}

	if fixtures {
		fakeFixture := fakeFootballAPIRepositoryFixture(matchID)
		apiFixtures = []repository.FootballApiFixture{fakeFixture}
	}

	return repository.Match{
		ID:                  matchID,
		HomeTeamID:          homeTeamID,
		AwayTeamID:          awayTeamID,
		StartsAt:            gofakeit.Date(),
		ResultStatus:        repository.Scheduled,
		FootballApiFixtures: apiFixtures,
		HomeTeam:            homeTeam,
		AwayTeam:            awayTeam,
	}
}

func fakeRepositoryTeam(teamID uint, aliases bool) repository.Team {
	var a []repository.Alias

	if aliases {
		fakeAlias := fakeRepositoryAlias(teamID)
		a = []repository.Alias{fakeAlias}
	}

	return repository.Team{
		ID:      teamID,
		Aliases: a,
	}
}

func fakeRepositoryAlias(teamID uint) repository.Alias {
	return repository.Alias{
		ID:              uint(gofakeit.Uint8()),
		TeamID:          teamID,
		Alias:           gofakeit.Name(),
		FootballApiTeam: nil,
	}
}

func fakeFootballAPIRepositoryFixture(matchID uint) repository.FootballApiFixture {
	data := pgtype.JSONB{}
	_ = data.UnmarshalJSON([]byte(footballAPIFixtureRaw(4, 2)))

	return repository.FootballApiFixture{
		ID:      uint(gofakeit.Uint8()),
		MatchID: matchID,
		Data:    data,
	}
}

func expectedMatch(repositoryMatch repository.Match) service.Match {
	var fixtures []service.FootballAPIFixture
	var homeTeam *service.Team
	var awayTeam *service.Team

	if repositoryMatch.HomeTeam != nil {
		var aliases []service.Alias

		if repositoryMatch.HomeTeam.Aliases != nil {
			aliases = make([]service.Alias, 0, len(repositoryMatch.HomeTeam.Aliases))
			for i := range repositoryMatch.HomeTeam.Aliases {
				aliases = append(aliases, service.Alias{
					Alias:  repositoryMatch.HomeTeam.Aliases[i].Alias,
					TeamID: repositoryMatch.HomeTeam.Aliases[i].TeamID,
				})
			}
		}

		homeTeam = &service.Team{
			ID:      repositoryMatch.HomeTeam.ID,
			Aliases: aliases,
		}
	}

	if repositoryMatch.AwayTeam != nil {
		var aliases []service.Alias

		if repositoryMatch.AwayTeam.Aliases != nil {
			aliases = make([]service.Alias, 0, len(repositoryMatch.AwayTeam.Aliases))
			for i := range repositoryMatch.AwayTeam.Aliases {
				aliases = append(aliases, service.Alias{
					Alias:  repositoryMatch.AwayTeam.Aliases[i].Alias,
					TeamID: repositoryMatch.AwayTeam.Aliases[i].TeamID,
				})
			}
		}

		awayTeam = &service.Team{
			ID:      repositoryMatch.AwayTeam.ID,
			Aliases: aliases,
		}
	}

	if repositoryMatch.FootballApiFixtures != nil {
		fixtures = make([]service.FootballAPIFixture, 0, len(repositoryMatch.FootballApiFixtures))
		for i := range repositoryMatch.FootballApiFixtures {
			fixtures = append(fixtures, service.FootballAPIFixture{
				ID:   repositoryMatch.FootballApiFixtures[i].ID,
				Home: 4,
				Away: 2,
			})
		}
	}

	return service.Match{
		ID:                  repositoryMatch.ID,
		StartsAt:            repositoryMatch.StartsAt,
		FootballApiFixtures: fixtures,
		HomeTeam:            homeTeam,
		AwayTeam:            awayTeam,
	}
}

func footballAPIFixtureRaw(home, away uint) string {
	return fmt.Sprintf(`{"goals": {"away": %d, "home": %d}, "teams": {"away": {"id": 35, "name": "Bournemouth"}, "home": {"id": 33, "name": "Manchester United"}}, "fixture": {"id": 1035330, "date": "2023-12-09T17:00:00+02:00", "status": {"long": "Match Finished", "short": "FT"}}}`, away, home)
}
