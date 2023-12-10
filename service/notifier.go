package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/andrewshostak/result-service/client"
	"github.com/andrewshostak/result-service/repository"
)

type NotifierService struct {
	subscriptionRepository SubscriptionRepository
	notifierClient         NotifierClient
}

func NewNotifierService(subscriptionRepository SubscriptionRepository, notifierClient NotifierClient) *NotifierService {
	return &NotifierService{subscriptionRepository: subscriptionRepository, notifierClient: notifierClient}
}

func (s *NotifierService) NotifySubscribers(ctx context.Context) error {
	subscriptions, err := s.subscriptionRepository.ListUnNotified(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("number of found subscriptions to notify: %d \n", len(subscriptions))

	mapped, err := fromRepositorySubscriptions(subscriptions)
	if err != nil {
		return fmt.Errorf("failed to map repository subscriptions: %w", err)
	}

	if len(mapped) == 0 {
		return nil
	}

	if mapped[0].Match == nil {
		return errors.New(fmt.Sprintf("match of the subscription %d is not found", mapped[0].ID))
	}

	if len(mapped[0].Match.FootballApiFixtures) == 0 {
		return errors.New(fmt.Sprintf("football api fixtures of the match with id %d is not found", mapped[0].MatchID))
	}

	for i := range subscriptions {
		notification := client.Notification{
			Url:  subscriptions[i].Url,
			Key:  subscriptions[i].Key,
			Home: mapped[0].Match.FootballApiFixtures[0].Home,
			Away: mapped[0].Match.FootballApiFixtures[0].Away,
		}

		toUpdate := repository.Subscription{Status: repository.SuccessfulSub}
		err := s.notifierClient.Notify(ctx, notification)
		if err != nil {
			fmt.Printf("failed to notify subscriber on url %s with the error: %s", subscriptions[i].Url, err.Error())
			toUpdate.Status = repository.ErrorSub
		}

		if toUpdate.Status == repository.SuccessfulSub {
			now := time.Now()
			toUpdate.NotifiedAt = &now
		}

		errUpdate := s.subscriptionRepository.Update(ctx, subscriptions[i].ID, toUpdate)
		if err != nil {
			return fmt.Errorf("failed to update subscription status to %s: %w", toUpdate.Status, errUpdate)
		}
	}

	return nil
}
