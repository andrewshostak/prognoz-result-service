package service

import (
	"context"

	"github.com/andrewshostak/result-service/repository"
)

type AliasRepository interface {
	Find(ctx context.Context, alias string) (*repository.Alias, error)
}
