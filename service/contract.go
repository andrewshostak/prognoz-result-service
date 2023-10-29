package service

import (
	"context"

	"github.com/andrewshostak/result-service/repository"
)

type AliasRepository interface {
	Find(ctx context.Context, alias string) (*repository.Alias, error)
}

type MatchRepository interface {
	Search(ctx context.Context, search repository.Match) (*repository.Match, error)
}
