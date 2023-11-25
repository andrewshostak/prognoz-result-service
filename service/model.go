package service

import (
	"time"

	"github.com/andrewshostak/result-service/client"
	"github.com/andrewshostak/result-service/repository"
)

type CreateMatchRequest struct {
	StartsAt  time.Time
	AliasHome string
	AliasAway string
}

type CreateSubscriptionRequest struct {
	MatchID   uint
	URL       string
	SecretKey string
}

type Match struct {
	ID       uint
	StartsAt time.Time
}

type Alias struct {
	Alias string
}

type FootballAPIFixture struct {
	ID    uint
	Match Match
}

type Result struct {
	Fixture Fixture
	Teams   Teams
	Goals   Goals
	Score   Score
}

type Fixture struct {
	ID     uint
	Status Status
	Date   string
}

type Teams struct {
	Home Team
	Away Team
}

type Team struct {
	ID   uint
	Name string
}

type Goals struct {
	Home uint
	Away uint
}

type Score struct {
	Fulltime  Goals
	Extratime Goals
}

type Status struct {
	Short string
}

func fromRepositoryFootballAPIFixture(f repository.FootballApiFixture) FootballAPIFixture {
	return FootballAPIFixture{
		ID:    f.ID,
		Match: fromRepositoryMatch(*f.Match),
	}
}

func fromClientFootballAPIFixture(c client.Result) Result {
	return Result{
		Fixture: Fixture{
			ID: c.Fixture.ID,
			Status: Status{
				Short: c.Fixture.Status.Short,
			},
			Date: c.Fixture.Date,
		},
		Teams: Teams{
			Home: Team{
				ID:   c.Teams.Home.ID,
				Name: c.Teams.Home.Name,
			},
			Away: Team{
				ID:   c.Teams.Away.ID,
				Name: c.Teams.Away.Name,
			},
		},
		Goals: Goals{
			Home: c.Goals.Home,
			Away: c.Goals.Away,
		},
		Score: Score{
			Fulltime: Goals{
				Home: c.Score.Fulltime.Home,
				Away: c.Score.Fulltime.Away,
			},
			Extratime: Goals{
				Home: c.Score.Extratime.Home,
				Away: c.Score.Extratime.Away,
			},
		},
	}
}

func fromRepositoryMatch(m repository.Match) Match {
	return Match{
		ID:       m.ID,
		StartsAt: m.StartsAt,
	}
}

func fromRepositoryAlias(a repository.Alias) Alias {
	return Alias{
		Alias: a.Alias,
	}
}
