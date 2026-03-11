package user

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func ById(ctx context.Context, db *sqlx.DB, userId string) (*User, error) {
	var u User
	err := db.GetContext(ctx, &u, `SELECT * FROM users WHERE user_id = ?`, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &u, nil
}

func ByEmailOrUsername(ctx context.Context, db *sqlx.DB, identifier string) (*User, error) {
	var u User
	err := db.GetContext(
		ctx,
		&u,
		`SELECT * FROM users WHERE email = ? OR username = ?`,
		identifier,
		identifier,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email or username: %w", err)
	}
	return &u, nil
}

func IsUsernameTaken(ctx context.Context, db *sqlx.DB, username string) (bool, error) {
	var exists bool
	err := db.GetContext(ctx, &exists,
		`SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)`,
		username)
	if err != nil {
		return false, fmt.Errorf("failed to check username: %w", err)
	}
	return exists, nil
}
