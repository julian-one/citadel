package session

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Delete removes a specific session (logout).
func Delete(ctx context.Context, db *sqlx.DB, sessionId string) error {
	_, err := db.ExecContext(ctx, `DELETE FROM sessions WHERE session_id = ?`, sessionId)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// DeleteAll removes all sessions for a user (logout everywhere).
func DeleteAll(ctx context.Context, db *sqlx.DB, userId string) error {
	_, err := db.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = ?`, userId)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}
	return nil
}
