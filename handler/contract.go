package handler

import (
	"context"

	"github.com/andrewshostak/result-service/service"
)

type MatchService interface {
	Create(ctx context.Context, request service.CreateMatchRequest) (string, error)
}
