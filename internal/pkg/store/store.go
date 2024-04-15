package store

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ougirez/diplom/internal/domain"
)

type Store interface {
	UserStore
	RegionItemStore
}

type store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) Store {
	return &store{pool}
}

type UserStore interface {
	CreateUser(ctx context.Context, user *domain.User) error
}
