package service

import "time"

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
