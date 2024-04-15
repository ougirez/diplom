package store

import (
	"context"
	"github.com/ougirez/diplom/internal/domain"
)

var userColumns = []string{"id", "email", "first_name", "last_name", "password_hash", "password_salt"}

func (s *store) CreateUser(ctx context.Context, user *domain.User) error {
	query := builder().Insert(tableUsers).
		Columns(userColumns[1:]...).
		Values(user.Email, user.FirstName, user.LastName, user.UserPassword.Hash, user.UserPassword.Salt).
		Suffix("RETURNING id")

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}
	if _, err = s.pool.Exec(ctx, sql, args); err != nil {
		return err
	}

	return nil
}
