package route

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"citadel/internal/middleware"
	"citadel/internal/session"
	"citadel/internal/user"

	"github.com/jmoiron/sqlx"
)

func ListUsers(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("ListUsers called")

		opts, err := user.ParseListOptions(r)
		if err != nil {
			logger.Error("Failed to parse user list options", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
			return
		}

		ctx := r.Context()
		users, err := user.List(ctx, db, opts)
		if err != nil {
			logger.Error("Failed to list users", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to list users"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(users)
	}
}

func GetUser(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("GetUser called")

		ctx := r.Context()
		userId := r.PathValue("id")
		logger.Info("Retrieving user", "user_id", userId)
		u, err := user.ById(ctx, db, userId)
		if err != nil {
			logger.Error("Failed to get user", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get user"})
			return
		}
		logger.Info("User retrieved successfully", "user_id", u.Id)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(u)
	}
}

func UpdateUser(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	type Request struct {
		Username *string `json:"username,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("UpdateUser called")

		var req Request
		if json.NewDecoder(r.Body).Decode(&req) != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		ctx := r.Context()
		taken, err := user.IsUsernameTaken(ctx, db, *req.Username)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to check username"})
			return
		}
		if taken {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Username is already taken"})
			return
		}

		s, ok := ctx.Value(middleware.SessionContextKey).(*session.Session)
		if !ok || s == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
			return
		}

		id := r.PathValue("id")
		if s.User != id {
			u, err := user.ById(ctx, db, s.User)
			if err != nil || u.Role != user.RoleAdmin {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).
					Encode(map[string]string{"error": "You can only update your own username"})
				return
			}
		}
		u, err := user.Update(ctx, db, id, req.Username, nil)
		if err != nil {
			logger.Error("Failed to update user", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update user role"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(u)
	}
}

func UpdateUserRole(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	type Request struct {
		Role string `json:"role"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("UpdateUserRole called")

		var req Request
		if json.NewDecoder(r.Body).Decode(&req) != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		// validate role
		if !user.Role(req.Role).Valid() {
			logger.Error("Invalid role provided", "role", req.Role)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid role provided"})
			return
		}

		ctx := r.Context()
		id := r.PathValue("id")
		_, err := user.Update(ctx, db, id, nil, &req.Role)
		if err != nil {
			logger.Error("Failed to update user role", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update user role"})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
