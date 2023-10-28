package service

import "time"

type CreateMatchRequest struct {
	StartsAt  time.Time
	AliasHome string
	AliasAway string
}
