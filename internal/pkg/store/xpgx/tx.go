package xpgx

import "github.com/jackc/pgx/v5"

// Tx - обертка над pgx.Tx с методами для работы с sqlizer.
type Tx interface {
	pgx.Tx
	Executor
}

type txImpl struct {
	pgx.Tx
	*executorImpl
}

// NewTx возвращает реализацию интерфейса Tx.
func NewTx(tx pgx.Tx) Tx {
	return &txImpl{tx, &executorImpl{tx}}
}
