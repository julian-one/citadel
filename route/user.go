package route

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"citadel/internal/user"

	"github.com/jmoiron/sqlx"
)

func ListUsers(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		users, err := user.List(ctx, db)
		if err != nil {
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
		Role     *string `json:"role,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req Request
		if json.NewDecoder(r.Body).Decode(&req) != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}
		// validate role, if provided
		if req.Role != nil && !user.Role(*req.Role).Valid() {
			logger.Error("Invalid role provided", "role", req.Role)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid role provided"})
			return
		}

		if req.Username != nil {
			taken, err := user.IsUsernameTaken(ctx, db, *req.Username)
			if err != nil {
				logger.Error("Failed to check username availability", "error", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).
					Encode(map[string]string{"error": "Failed to check username availability"})
				return
			}
			if taken {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Username is already taken"})
				return
			}
		}

		id := r.PathValue("id")
		u, err := user.Update(ctx, db, id, req.Username, req.Role)
		if err != nil {
			logger.Error("Failed to update user role", "error", err)
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
