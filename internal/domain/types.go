package domain

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type NullString = pgtype.Text

func NewNullString(s string) NullString {
	return pgtype.Text{String: s, Valid: len(s) > 0}
}
