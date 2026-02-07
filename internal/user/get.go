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
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &u, nil
}

func ByEmail(ctx context.Context, db *sqlx.DB, email string) (*User, error) {
	var u User
	err := db.GetContext(ctx, &u, `SELECT * FROM users WHERE email = ?`, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
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
