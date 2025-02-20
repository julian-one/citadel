package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"citadel/internal/user"

	"github.com/jmoiron/sqlx"
)

const (
	SessionDuration = 24 * time.Hour
	SessionIDLength = 32 // 32 bytes = 64 hex chars
)

const (
	CookieName     = "session_id"
	cookieMaxAge   = int(24 * time.Hour / time.Second) // 24 hours in seconds
	cookiePath     = "/"
	cookieSecure   = false // Set to true in production with HTTPS
	cookieHTTPOnly = true
	cookieSameSite = http.SameSiteStrictMode
)

func SetSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    sessionID,
		Path:     cookiePath,
		MaxAge:   cookieMaxAge,
		Secure:   cookieSecure,
		HttpOnly: cookieHTTPOnly,
		SameSite: cookieSameSite,
	})
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     cookiePath,
		MaxAge:   -1, // Immediately expire
		Secure:   cookieSecure,
		HttpOnly: cookieHTTPOnly,
		SameSite: cookieSameSite,
	})
}

// Create generates a new session for the user.
func Create(
	ctx context.Context,
	db *sqlx.DB,
	userId int64,
) (string, error) {
	sessionId, err := generateSessionId()
	if err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}

	expiresAt := time.Now().Add(SessionDuration)

	_, err = db.ExecContext(ctx, `
		INSERT INTO sessions (session_id, user_id, expires_at)
		VALUES (?, ?, ?)
	`, sessionId, userId, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionId, nil
}

// Validate checks if session exists and is not expired, returns user data.
func Validate(ctx context.Context, db *sqlx.DB, sessionID string) (*user.User, error) {
	var u user.User
	err := db.GetContext(ctx, &u, `
		SELECT u.*
		FROM sessions s
		JOIN users u ON s.user_id = u.user_id
		WHERE s.session_id = ? AND s.expires_at > datetime('now')
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired session: %w", err)
	}
	return &u, nil
}

// Delete removes a specific session (logout).
func Delete(ctx context.Context, db *sqlx.DB, sessionID string) error {
	_, err := db.ExecContext(ctx, `DELETE FROM sessions WHERE session_id = ?`, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// DeleteAllForUser removes all sessions for a user (logout everywhere).
func DeleteAllForUser(ctx context.Context, db *sqlx.DB, userID int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}
	return nil
}

// CleanupExpired removes expired sessions (call periodically).
func CleanupExpired(ctx context.Context, db *sqlx.DB) (int64, error) {
	result, err := db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at <= datetime('now')`)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}
	return result.RowsAffected()
}

func generateSessionId() (string, error) {
	bytes := make([]byte, SessionIDLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
