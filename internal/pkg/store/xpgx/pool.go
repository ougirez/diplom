package xpgx

import "github.com/jackc/pgx/v5/pgxpool"

// Pool - обертка над types.Pool с методами для работы с sqlizer.
type Pool interface {
	Executor
}

type poolImpl struct {
	*pgxpool.Pool
	*executorImpl
}

// NewPool возвращает реализацию интерфейса Pool.
func NewPool(pool *pgxpool.Pool) Pool {
	return &poolImpl{pool, &executorImpl{pool}}
}
