package route

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	recipereview "citadel/internal/recipe/review"
	"citadel/internal/session"

	"github.com/jmoiron/sqlx"
)

func CreateRecipeReview(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		s, ok := ctx.Value(session.ContextKey).(*session.Session)
		if !ok || s == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
			return
		}

		recipeID := r.PathValue("id")

		var req recipereview.CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn("failed to decode create recipe review request", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		req.User = s.User
		req.Recipe = recipeID

		tx, err := db.BeginTxx(ctx, &sql.TxOptions{})
		if err != nil {
			logger.Error("failed to begin transaction", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create recipe review"})
			return
		}
		defer tx.Rollback()

		reviewID, err := recipereview.Create(ctx, tx, req)
		if err != nil {
			if errors.Is(err, recipereview.ErrDuplicateReview) ||
				strings.Contains(err.Error(), "rating must be") ||
				strings.Contains(err.Error(), "difficulty must be") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			logger.Error("failed to create recipe review", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create recipe review"})
			return
		}

		if err := tx.Commit(); err != nil {
			logger.Error("failed to commit transaction", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create recipe review"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"review_id": reviewID})
	}
}

func ListRecipeReviews(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		recipeID := r.PathValue("id")

		reviews, err := recipereview.ByRecipe(ctx, db, recipeID)
		if err != nil {
			logger.Error("failed to list recipe reviews", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to list recipe reviews"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(reviews)
	}
}

func DeleteRecipeReview(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		s, ok := ctx.Value(session.ContextKey).(*session.Session)
		if !ok || s == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
			return
		}

		reviewID := r.PathValue("id")

		err := recipereview.Delete(ctx, db, s.User, reviewID)
		if err != nil {
			if strings.Contains(err.Error(), "not found or not owned") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]string{"error": "Forbidden"})
				return
			}
			logger.Error("failed to delete recipe review", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete recipe review"})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
