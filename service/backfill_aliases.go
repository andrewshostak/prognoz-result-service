package service

import (
	"context"
	"fmt"
)

type BackfillAliasesService struct {
	aliasRepository   AliasRepository
	footballAPIClient FootballAPIClient
	seasonHelper      SeasonHelper
	logger            Logger
}

func NewBackfillAliasesService(
	aliasRepository AliasRepository,
	footballAPIClient FootballAPIClient,
	seasonHelper SeasonHelper,
	logger Logger,
) *BackfillAliasesService {
	return &BackfillAliasesService{
		aliasRepository:   aliasRepository,
		footballAPIClient: footballAPIClient,
		seasonHelper:      seasonHelper,
		logger:            logger,
	}
}

func (s *BackfillAliasesService) Backfill(ctx context.Context) error {
	s.logger.Info().Msg("starting aliases backfill")

	season := s.seasonHelper.CurrentSeason()

	s.logger.Info().Int("season", season).Msg("searching leagues")

	result, err := s.footballAPIClient.SearchLeagues(ctx, uint(season))
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

	return nil
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
		{name: "Süper Lig", country: "Turkey", euroCupsIntersected: true},
		{name: "Premiership", country: "Scotland", euroCupsIntersected: true},
		{name: "Czech Liga", country: "Czech-Republic", euroCupsIntersected: true},
		{name: "Super League", country: "Switzerland", euroCupsIntersected: true},
		{name: "Bundesliga", country: "Austria", euroCupsIntersected: true},
		{name: "Superliga", country: "Denmark", euroCupsIntersected: true},
		{name: "Eliteserien", country: "Norway", euroCupsIntersected: true},
		{name: "Ligat Ha'al", country: "Israel", euroCupsIntersected: true},
		{name: "Super League 1", country: "Greece", euroCupsIntersected: true},
		{name: "Super Liga", country: "Serbia", euroCupsIntersected: true},
		{name: "Ekstraklasa", country: "Poland", euroCupsIntersected: true},
		{name: "HNL", country: "Croatia", euroCupsIntersected: true},
	}
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
	name                string
	country             string
	euroCupsIntersected bool
}
