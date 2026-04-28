package session

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// ById retrieves session by id
func ById(ctx context.Context, db sqlx.QueryerContext, sessionId string) (*Session, error) {
	var s Session
	err := sqlx.GetContext(ctx, db, &s,
		`SELECT * FROM sessions WHERE session_id = ? AND expires_at > datetime('now')`,
		sessionId)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &s, nil
}

// IsValid checks if the session is valid (exists and not expired)
func IsValid(ctx context.Context, db sqlx.QueryerContext, sessionId string) error {
	var exists bool
	err := sqlx.GetContext(ctx, db, &exists,
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
