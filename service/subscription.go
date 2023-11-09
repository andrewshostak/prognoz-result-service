package service

import (
	"context"
	"fmt"
	"time"

	"github.com/andrewshostak/result-service/repository"
)

type SubscriptionService struct {
	subscriptionRepository SubscriptionRepository
}

func NewSubscriptionService(subscriptionRepository SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{subscriptionRepository: subscriptionRepository}
}

func (s *SubscriptionService) Create(ctx context.Context, request CreateSubscriptionRequest) error {
	_, err := s.subscriptionRepository.Create(ctx, repository.Subscription{
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
