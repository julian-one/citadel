package pokemon

import (
	"context"

	"github.com/jmoiron/sqlx"
)

func ByName(ctx context.Context, db sqlx.QueryerContext, name string) (*Pokemon, error) {
	var p Pokemon
	err := sqlx.GetContext(ctx, db,
		&p,
		`SELECT pokemon_id, name, height, weight FROM pokemon WHERE name = ?`,
		name,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func Exists(ctx context.Context, db sqlx.QueryerContext, name string) bool {
	var exists bool
	err := db.QueryRowxContext(ctx, `SELECT EXISTS(SELECT 1 FROM pokemon WHERE name = ?)`, name).
		Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}
