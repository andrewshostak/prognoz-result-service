package service

import "context"

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
	s.logger.Info().Msg("hello world")

	_ = s.seasonHelper.CurrentSeason()

	return nil
}

func (s *BackfillAliasesService) leagues() []league {
	return []league{
		{name: "Premier League", country: "Ukraine"},
		{name: "Premier League", country: "England"},
		{name: "La Liga", country: "Spain"},
		// TODO: add more leagues
	}
}

type league struct {
	name    string
	country string
}
