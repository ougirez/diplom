package store

import (
	"errors"
	"github.com/ougirez/diplom/internal/pkg/constants"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

const (
	tableUsers       = "users"
	tableRegionItems = "region_items"
)

var mapping = map[error]error{pgx.ErrNoRows: constants.ErrDBNotFound}

func wrapErr(err error) error {
	for k, v := range mapping {
		if errors.Is(err, k) {
			return v
		}
	}
	return err
}

// builder возвращает squirrel SQL Builder обьект.
func builder() squirrel.StatementBuilderType {
	return squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
}
