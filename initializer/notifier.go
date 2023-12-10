package initializer

import (
	"context"
	"time"
)

const checkSubscribersTime = 1 * time.Minute

type NotifierInitializer struct {
	notifierService NotifierService
}

func NewNotifierInitializer(notifierService NotifierService) *NotifierInitializer {
	return &NotifierInitializer{notifierService: notifierService}
}

func (i *NotifierInitializer) Start() {
	ticker := time.NewTicker(checkSubscribersTime)

	go func() {
		for {
			select {
			case <-ticker.C:
				ctx := context.Background()
				err := i.notifierService.NotifySubscribers(ctx)
				if err != nil {
					panic(err)
				}
			}
		}
	}()
}
