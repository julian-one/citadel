package session

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	SessionDuration = 24 * time.Hour
	SessionIdLength = 32 // 32 bytes = 64 hex chars
)

const (
	CookieName     = "TOKEN"
	cookieMaxAge   = int(24 * time.Hour / time.Second) // 24 hours in seconds
	cookiePath     = "/"
	cookieSecure   = false // TODO: Set to true in production with HTTPS
	cookieHTTPOnly = true
	cookieSameSite = http.SameSiteLaxMode
)

type Session struct {
	SessionId string    `json:"session_id" db:"session_id"`
	User      string    `json:"user_id"    db:"user_id"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func SetSessionCookie(w http.ResponseWriter, sessionId string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    sessionId,
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

func Create(
	ctx context.Context,
	db *sqlx.DB,
	userId string,
) (*Session, error) {
	// define session expiration time
	expiresAt := time.Now().Add(SessionDuration)

	// generate a secure random session id
	sessionId := uuid.New().String()

	// insert session into database and return the created session
	var s Session
	err := db.QueryRowxContext(ctx,
		`INSERT INTO sessions (session_id, user_id, expires_at) VALUES (?, ?, ?) RETURNING *;`,
		sessionId, userId, expiresAt,
	).StructScan(&s)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	return &s, nil
}

// IsValid checks if session exists and is not expired
func IsValid(ctx context.Context, db *sqlx.DB, sessionId string) (bool, error) {
	var exists bool
	err := db.GetContext(ctx, &exists,
		`SELECT EXISTS(
			SELECT 1 FROM sessions 
			WHERE session_id = ? AND expires_at > datetime('now')
		)`,
		sessionId)
	if err != nil {
		return false, fmt.Errorf("failed to validate session: %w", err)
	}
	return exists, nil
}

// Get retrieves session by id
func Get(ctx context.Context, db *sqlx.DB, sessionId string) (*Session, error) {
	var s Session
	err := db.GetContext(ctx, &s,
		`SELECT * FROM sessions WHERE session_id = ? AND expires_at > datetime('now')`,
		sessionId)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &s, nil
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
