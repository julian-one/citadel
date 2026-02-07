package route

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"citadel/internal/session"
	"citadel/internal/user"

	"github.com/jmoiron/sqlx"
)

func Register(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	type Request struct {
		Username string `json:"username"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("handling register request")
		ctx := r.Context()

		logger.Info("before basic auth")
		email, password, ok := r.BasicAuth()
		if !ok {
			logger.Info("invalid basic auth credentials")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid basic auth credentials"})
			return
		}
		logger.Info("received register request", slog.String("email", email))

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Info("failed to decode register request body", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}
		logger.Info("decoded register request body", slog.String("username", req.Username))

		userId, err := user.Create(ctx, db, user.CreateRequest{
			Username: req.Username,
			Email:    email,
			Password: password,
		})
		if err != nil {
			logger.Info("failed to create user", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create user"})
			return
		}
		logger.Info("created user", slog.String("userId", userId))

		s, err := session.Create(
			ctx,
			db,
			userId,
		)
		if err != nil {
			logger.Info("failed to create session", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create session"})
			return
		}
		session.SetSessionCookie(w, s.SessionId)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(s)
	}
}

func Login(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	type Request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		logger.Info("handling login request")
		// logger.Info("request headers", slog.Any("headers", r.Header))
		email, password, ok := r.BasicAuth()
		if !ok {
			logger.Info("invalid basic auth credentials")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid basic auth credentials"})
			return
		}
		logger.Info("received login request", slog.String("email", email))

		u, err := user.ByEmail(ctx, db, email)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid email or password"})
			return
		}

		match, err := user.Verify(password, u.Hash, u.Salt)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Authentication error"})
			return
		}
		if !match {
			logger.Info("invalid password for email", slog.String("email", email))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid email or password"})
			return
		}

		s, err := session.Create(ctx, db, u.Id)
		if err != nil {
			logger.Error("failed to create session", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create session"})
			return
		}
		session.SetSessionCookie(w, s.SessionId)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(s)
	}
}

func Logout(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("handling logout request")

		cookie, err := r.Cookie(session.CookieName)
		if err != nil || cookie.Value == "" {
			// don't return an error if the cookie is missing or empty, just treat it as a successful logout
			logger.Info("no session cookie found, treating as successful logout")
		} else {
			err = session.Delete(r.Context(), db, cookie.Value)
			if err != nil {
				// don't return an error if session deletion fails, just log it and continue with clearing the cookie
				logger.Error("failed to delete session", "error", err)
			}
		}

		session.ClearSessionCookie(w)
		w.WriteHeader(http.StatusNoContent)
	}
}
