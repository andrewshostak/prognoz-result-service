package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/andrewshostak/result-service/errs"
	"github.com/andrewshostak/result-service/repository"
)

type MatchService struct {
	aliasRepository AliasRepository
	matchRepository MatchRepository
}

func NewMatchService(aliasRepository AliasRepository, matchRepository MatchRepository) *MatchService {
	return &MatchService{aliasRepository: aliasRepository, matchRepository: matchRepository}
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

	return 123, nil
}
