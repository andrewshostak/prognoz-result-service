package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/andrewshostak/result-service/client"
)

type BackfillAliasesService struct {
	aliasRepository   AliasRepository
	footballAPIClient FootballAPIClient
	logger            Logger
}

func NewBackfillAliasesService(
	aliasRepository AliasRepository,
	footballAPIClient FootballAPIClient,
	logger Logger,
) *BackfillAliasesService {
	return &BackfillAliasesService{
		aliasRepository:   aliasRepository,
		footballAPIClient: footballAPIClient,
		logger:            logger,
	}
}

func (s *BackfillAliasesService) Backfill(ctx context.Context, season uint) error {
	s.logger.Info().Msg("starting aliases backfill")
	s.logger.Info().Uint("season", season).Msg("searching leagues")

	result, err := s.footballAPIClient.SearchLeagues(ctx, season)
	if err != nil {
		return fmt.Errorf("failed to search leagues: %w", err)
	}

	s.logger.Info().Int("length", len(result.Response)).Msg("leagues found")

	allLeagues := make([]LeagueData, 0, len(result.Response))
	for i := range result.Response {
		allLeagues = append(allLeagues, fromClientFootballAPILeague(result.Response[i]))
	}

	leagues := s.filterOutLeagues(allLeagues, s.getIncludedLeagues())

	s.logger.Info().Int("length", len(leagues)).Msg("leagues filtering is done")

	leaguesTeams, err := s.getLeaguesTeams(ctx, leagues, season)
	if err != nil {
		return fmt.Errorf("failed to get teams: %w", err)
	}

	s.saveTeams(ctx, leaguesTeams)

	return nil
}

func (s *BackfillAliasesService) getLeaguesTeams(ctx context.Context, leagues []LeagueData, season uint) (map[LeagueData][]TeamExternal, error) {
	const numberOfWorkers = 3
	jobs := make(chan struct{}, numberOfWorkers)
	wg := sync.WaitGroup{}
	var mutex = &sync.RWMutex{}

	teams := map[LeagueData][]TeamExternal{}

	for i := range leagues {
		wg.Add(1)
		jobs <- struct{}{}

		s.logger.Info().
			Int("number", i).
			Str("league_name", leagues[i].League.Name).
			Str("country_name", leagues[i].Country.Name).
			Msg("iteration")

		go func(ctx context.Context, league LeagueData) {
			result, err := s.footballAPIClient.SearchTeams(ctx, client.TeamsSearch{Season: season, League: league.League.ID})
			<-jobs

			if err != nil {
				s.logger.Error().Err(err).
					Str("league_name", league.League.Name).
					Str("country_name", league.Country.Name).
					Msg("failed to get teams")
				return
			}

			s.logger.Info().
				Str("league_name", league.League.Name).
				Str("country_name", league.Country.Name).
				Msg("successfully get teams")

			mutex.Lock()
			teams[league] = fromClientFootballAPITeams(result.Response)
			mutex.Unlock()

			defer wg.Done()
		}(ctx, leagues[i])
	}

	wg.Wait()

	s.logger.Info().Int("number", len(teams)).Msg("received results")

	return teams, nil
}

func (s *BackfillAliasesService) filterOutLeagues(allLeagues []LeagueData, includedLeagues []league) []LeagueData {
	filtered := make([]LeagueData, 0, len(includedLeagues))
	for i := range allLeagues {
		if isIncludedLeague(allLeagues[i], includedLeagues) {
			filtered = append(filtered, allLeagues[i])
		}
	}

	return filtered
}

func (s *BackfillAliasesService) getIncludedLeagues() []league {
	return []league{
		// all
		// european cups
		{name: "UEFA Champions League", country: "World"},
		{name: "UEFA Europa League", country: "World"},
		{name: "UEFA Europa Conference League", country: "World"},
		// national teams competitions
		{name: "Euro Championship - Qualification", country: "World"},
		{name: "Euro Championship", country: "World"},
		{name: "World Cup - Qualification South America", country: "World"},
		{name: "World Cup", country: "World"},
		{name: "Copa America", country: "World"},
		{name: "Africa Cup of Nations", country: "World"},
		// top leagues + ukrainian league
		{name: "Premier League", country: "Ukraine"},
		{name: "Premier League", country: "England"},
		{name: "La Liga", country: "Spain"},
		{name: "Serie A", country: "Italy"},
		{name: "Bundesliga", country: "Germany"},
		{name: "Ligue 1", country: "France"},
		{name: "Eredivisie", country: "Netherlands"},
		{name: "Primeira Liga", country: "Portugal"},
		{name: "Jupiler Pro League", country: "Belgium"},
		// only intersected with euro cups: Champions/Europa/Conference League
		//{name: "Süper Lig", country: "Turkey"},
		//{name: "Premiership", country: "Scotland"},
		//{name: "Czech Liga", country: "Czech-Republic"},
		//{name: "Super League", country: "Switzerland"},
		//{name: "Bundesliga", country: "Austria"},
		//{name: "Superliga", country: "Denmark"},
		//{name: "Eliteserien", country: "Norway"},
		//{name: "Ligat Ha'al", country: "Israel"},
		//{name: "Super League 1", country: "Greece"},
		//{name: "Super Liga", country: "Serbia"},
		//{name: "Ekstraklasa", country: "Poland"},
		//{name: "HNL", country: "Croatia"},
	}
}

func (s *BackfillAliasesService) saveTeams(ctx context.Context, leaguesTeams map[LeagueData][]TeamExternal) {
	const numberOfWorkers = 3
	jobs := make(chan struct{}, numberOfWorkers)
	wg := sync.WaitGroup{}

	for league, teams := range leaguesTeams {
		wg.Add(1)
		jobs <- struct{}{}

		go func(league LeagueData, teams []TeamExternal) {
			numberOfSaved, numberOfExisted := 0, 0
			for i := range teams {
				_, err := s.aliasRepository.Find(ctx, teams[i].Name)
				if err == nil {
					s.logger.Info().
						Str("alias", teams[i].Name).
						Uint("football_api_team_id", teams[i].ID).
						Msg("alias already exists")
					numberOfExisted++
					continue
				}

				errTrx := s.aliasRepository.SaveInTrx(ctx, teams[i].Name, teams[i].ID)
				if errTrx != nil {
					s.logger.Error().
						Str("alias", teams[i].Name).
						Uint("football_api_team_id", teams[i].ID).
						Err(errTrx).
						Msg("failed to save alias")
					continue
				}
				numberOfSaved++
			}

			s.logger.Info().
				Str("league_name", league.League.Name).
				Str("country_name", league.Country.Name).
				Int("number_of_saved", numberOfSaved).
				Int("number_of_existed", numberOfExisted).
				Msg("league teams saving finished")

			defer wg.Done()
		}(league, teams)
	}

	wg.Wait()
}

func isIncludedLeague(league LeagueData, includedLeagues []league) bool {
	for i := range includedLeagues {
		if includedLeagues[i].name == league.League.Name && includedLeagues[i].country == league.Country.Name {
			return true
		}
	}

	return false
}

type league struct {
	name    string
	country string
}
