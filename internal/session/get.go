package session

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// ById retrieves session by id
func ById(ctx context.Context, db *sqlx.DB, sessionId string) (*Session, error) {
	var s Session
	err := db.GetContext(ctx, &s,
		`SELECT * FROM sessions WHERE session_id = ? AND expires_at > datetime('now')`,
		sessionId)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &s, nil
}

// IsValid checks if the session is valid (exists and not expired)
func IsValid(ctx context.Context, db *sqlx.DB, sessionId string) error {
	var exists bool
	err := db.GetContext(ctx, &exists,
		`SELECT EXISTS(
			SELECT 1 FROM sessions 
			WHERE session_id = ? AND expires_at > datetime('now')
		)`,
		sessionId)
	if err != nil {
		return fmt.Errorf("failed to validate session: %w", err)
	}
	if !exists {
		return fmt.Errorf("invalid or expired session")
	}
	return nil
}
