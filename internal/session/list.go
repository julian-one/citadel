package session

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

func List(ctx context.Context, db *sqlx.DB, userId string) ([]Session, error) {
	var s []Session

	query, args, err := sq.Select("*").
		From("sessions").
		Where(sq.Eq{"user_id": userId}).
		OrderBy("expires_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	err = db.SelectContext(ctx, &s, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list session: %w", err)
	}

	return s, nil
}
