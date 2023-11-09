package handler

import (
	"time"

	"github.com/andrewshostak/result-service/service"
)

type CreateMatchRequest struct {
	StartsAt  time.Time `binding:"required" json:"starts_at" time_format:"2006-01-02T15:04:05Z"`
	AliasHome string    `binding:"required" json:"alias_home"`
	AliasAway string    `binding:"required" json:"alias_away"`
}

type CreateSubscriptionRequest struct {
	MatchID   uint   `binding:"required" json:"match_id"`
	URL       string `binding:"required" json:"url"`
	SecretKey string `binding:"required" json:"secret_key"`
}

func (cmr *CreateMatchRequest) ToDomain() service.CreateMatchRequest {
	return service.CreateMatchRequest{
		StartsAt:  cmr.StartsAt,
		AliasHome: cmr.AliasHome,
		AliasAway: cmr.AliasAway,
	}
}

func (csr *CreateSubscriptionRequest) ToDomain() service.CreateSubscriptionRequest {
	return service.CreateSubscriptionRequest{
		MatchID:   csr.MatchID,
		URL:       csr.URL,
		SecretKey: csr.SecretKey,
	}
}
