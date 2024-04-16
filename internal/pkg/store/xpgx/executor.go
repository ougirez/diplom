package xpgx

import (
	"context"
	"errors"

	"github.com/georgysavva/scany/v2/dbscan"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// AllowUnknownColumns позволяет сканировать строки в структуры,
// в которых отсутствует часть колонок из строк.
func AllowUnknownColumns() {
	dbscan.WithAllowUnknownColumns(true)(dbscan.DefaultAPI)
	pgxscan.DefaultAPI, _ = pgxscan.NewAPI(dbscan.DefaultAPI)
}

// BasicExecutor - интерфейс с базовыми методами, на которых строятся операции интерфейса Executor.
type BasicExecutor interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Sqlizer описывает метод ToSql, возвращающий SQL и аргументы для работы с БД.
// Предназначен для интеграции с https://pkg.go.dev/github.com/Masterminds/squirrel#Sqlizer.
type Sqlizer interface {
	ToSql() (sql string, args []interface{}, err error)
}

// Executor - расширяет BasicExecutor методами для работы с Sqlizer.
type Executor interface {
	BasicExecutor

	Execx(ctx context.Context, sqlizer Sqlizer) (pgconn.CommandTag, error)
	Queryx(ctx context.Context, sqlizer Sqlizer) (pgx.Rows, error)
	QueryxRow(ctx context.Context, sqlizer Sqlizer) pgx.Row

	Get(ctx context.Context, destPtr interface{}, sql string, args ...interface{}) error
	Getx(ctx context.Context, destPtr interface{}, sqlizer Sqlizer) error

	Select(ctx context.Context, destSlice interface{}, sql string, args ...interface{}) error
	Selectx(ctx context.Context, destSlice interface{}, sqlizer Sqlizer) error
}

type executorImpl struct {
	BasicExecutor
}

// Execx выполняет запрос с переданным sqlizer.
func (e *executorImpl) Execx(ctx context.Context, sqlizer Sqlizer) (pgconn.CommandTag, error) {
	sql, args, err := sqlizer.ToSql()
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	return e.Exec(ctx, sql, args...)
}

// Queryx выполняет запрос с переданным sqlizer.
func (e *executorImpl) Queryx(ctx context.Context, sqlizer Sqlizer) (pgx.Rows, error) {
	sql, args, err := sqlizer.ToSql()
	if err != nil {
		return nil, err
	}
	return e.Query(ctx, sql, args...)
}

// QueryxRow выполняет запрос с переданным sqlizer.
func (e *executorImpl) QueryxRow(ctx context.Context, sqlizer Sqlizer) pgx.Row {
	sql, args, err := sqlizer.ToSql()
	if err != nil {
		return errRow{err}
	}
	return e.QueryRow(ctx, sql, args...)
}

// Get выполняет запрос и возвращает в destPtr одну запись. Get == QueryRow.Scan.
func (e *executorImpl) Get(ctx context.Context, destPtr interface{}, sql string, args ...interface{}) error {
	return pgxscan.Get(ctx, e, destPtr, sql, args...)
}

// Getx выполняет запрос с переданным sqlizer и возвращает в destPtr одну запись. Get == QueryRow.Scan.
func (e *executorImpl) Getx(ctx context.Context, destPtr interface{}, sqlizer Sqlizer) error {
	sql, args, err := sqlizer.ToSql()
	if err != nil {
		return err
	}
	return e.Get(ctx, destPtr, sql, args...)
}

// Select выполняет запрос и возвращает в destSlice массив результатов. Никогда не возвращает ErrNoRows.
func (e *executorImpl) Select(ctx context.Context, destSlice interface{}, sql string, args ...interface{}) error {
	err := pgxscan.Select(ctx, e, destSlice, sql, args...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	return err
}

// Selectx выполняет запрос с переданным sqlizer и возвращает в destSlice массив результатов. Никогда не возвращает ErrNoRows.
func (e *executorImpl) Selectx(ctx context.Context, destSlice interface{}, sqlizer Sqlizer) error {
	sql, args, err := sqlizer.ToSql()
	if err != nil {
		return err
	}
	return e.Select(ctx, destSlice, sql, args...)
}

type errRow struct {
	error
}

func (er errRow) Scan(...interface{}) error {
	return er.error
}
