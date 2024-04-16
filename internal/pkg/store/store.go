package store

import (
	"context"
	"github.com/ougirez/diplom/internal/domain"
	"github.com/ougirez/diplom/internal/pkg/store/xpgx"
)

type Tx = xpgx.Tx
type Pool = xpgx.Pool

type Store interface {
	UserStore
	ProviderStore
}

type store struct {
	pool Pool
}

func NewStore(pool Pool) Store {
	return &store{pool}
}

type UserStore interface {
	CreateUser(ctx context.Context, user *domain.User) error
}
