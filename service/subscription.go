package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/andrewshostak/result-service/repository"
)

type SubscriptionService struct {
	subscriptionRepository SubscriptionRepository
	matchRepository        MatchRepository
}

func NewSubscriptionService(subscriptionRepository SubscriptionRepository, matchRepository MatchRepository) *SubscriptionService {
	return &SubscriptionService{subscriptionRepository: subscriptionRepository, matchRepository: matchRepository}
}

func (s *SubscriptionService) Create(ctx context.Context, request CreateSubscriptionRequest) error {
	match, err := s.matchRepository.One(ctx, repository.Match{ID: request.MatchID})
	if err != nil {
		return fmt.Errorf("failed to get a match: %w", err)
	}

	if match.ResultStatus != "scheduled" {
		return errors.New("match status is not scheduled")
	}

	_, err = s.subscriptionRepository.Create(ctx, repository.Subscription{
		MatchID:   request.MatchID,
		Key:       request.SecretKey,
		CreatedAt: time.Now(),
		Url:       request.URL,
	})

	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}
