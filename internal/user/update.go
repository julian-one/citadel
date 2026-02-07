package user

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

func Update(
	ctx context.Context,
	db *sqlx.DB,
	userId string,
	username *string,
	role *string,
) (*User, error) {
	query := sq.Update("users").
		Set("updated_at", sq.Expr("datetime('now')")).
		Where(sq.Eq{"user_id": userId})

	if username != nil {
		query = query.Set("username", *username)
	}
	if role != nil {
		query = query.Set("role", *role)
	}

	sql, args, err := query.
		Suffix("RETURNING *").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	var user User
	err = db.QueryRowxContext(ctx, sql, args...).StructScan(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
