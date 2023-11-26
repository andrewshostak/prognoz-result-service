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

	FootballApiFixtures []FootballAPIFixture
	HomeTeam            *Team
	AwayTeam            *Team
}

type Team struct {
	ID uint

	Aliases []Alias
}

type Alias struct {
	Alias  string
	TeamID uint

	FootballApiTeam *FootballApiTeam
}

type FootballApiTeam struct {
	ID     uint
	TeamID uint
}

type FootballAPIFixture struct {
	ID uint
}

type Result struct {
	Fixture Fixture
	Teams   TeamsExternal
	Goals   Goals
	Score   Score
}

type Fixture struct {
	ID     uint
	Status Status
	Date   string
}

type TeamsExternal struct {
	Home TeamExternal
	Away TeamExternal
}

type TeamExternal struct {
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
		ID: f.ID,
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
		Teams: TeamsExternal{
			Home: TeamExternal{
				ID:   c.Teams.Home.ID,
				Name: c.Teams.Home.Name,
			},
			Away: TeamExternal{
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
	fixtures := make([]FootballAPIFixture, 0, len(m.FootballApiFixtures))
	for _, fixture := range m.FootballApiFixtures {
		fixtures = append(fixtures, fromRepositoryFootballAPIFixture(fixture))
	}

	var homeTeam *Team
	if m.HomeTeam != nil {
		aliases := make([]Alias, 0, len(m.HomeTeam.Aliases))
		for _, alias := range m.HomeTeam.Aliases {
			aliases = append(aliases, Alias{Alias: alias.Alias})
		}

		homeTeam = &Team{ID: m.HomeTeam.ID, Aliases: aliases}
	}

	var awayTeam *Team
	if m.AwayTeam != nil {
		aliases := make([]Alias, 0, len(m.AwayTeam.Aliases))
		for _, alias := range m.AwayTeam.Aliases {
			aliases = append(aliases, Alias{Alias: alias.Alias})
		}

		awayTeam = &Team{ID: m.AwayTeam.ID, Aliases: aliases}
	}
	return Match{
		ID:                  m.ID,
		StartsAt:            m.StartsAt,
		FootballApiFixtures: fixtures,
		HomeTeam:            homeTeam,
		AwayTeam:            awayTeam,
	}
}

func fromRepositoryMatches(m []repository.Match) []Match {
	matches := make([]Match, 0, len(m))
	for i := range m {
		matches = append(matches, fromRepositoryMatch(m[i]))
	}

	return matches
}

func fromRepositoryFootballAPITeam(t repository.FootballApiTeam) FootballApiTeam {
	return FootballApiTeam{
		ID:     t.ID,
		TeamID: t.TeamID,
	}
}

func fromRepositoryAlias(a repository.Alias) Alias {
	var footballAPITeam *FootballApiTeam

	if a.FootballApiTeam != nil {
		mapped := fromRepositoryFootballAPITeam(*a.FootballApiTeam)
		footballAPITeam = &mapped
	}

	return Alias{
		Alias:           a.Alias,
		TeamID:          a.TeamID,
		FootballApiTeam: footballAPITeam,
	}
}
