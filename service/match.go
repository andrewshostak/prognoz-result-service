package service

import (
	"context"
	"fmt"
)

type MatchService struct {
	aliasRepository AliasRepository
}

func NewMatchService(aliasRepository AliasRepository) *MatchService {
	return &MatchService{aliasRepository: aliasRepository}
}

func (s *MatchService) Create(ctx context.Context, request CreateMatchRequest) (string, error) {
	aliasHome, err := s.aliasRepository.Find(ctx, request.AliasHome)
	if err != nil {
		return "", fmt.Errorf("failed to find home alias: %w", err)
	}

	aliasAway, err := s.aliasRepository.Find(ctx, request.AliasAway)
	if err != nil {
		return "", fmt.Errorf("failed to find away alias: %w", err)
	}

	fmt.Printf("aliasHome: %v \n", aliasHome)
	fmt.Printf("aliasAway: %v \n", aliasAway)

	return "", nil
}
