package route

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"citadel/internal/middleware"
	recipebookmark "citadel/internal/recipe/bookmark"
	"citadel/internal/session"

	"github.com/jmoiron/sqlx"
)

func ListBookmarkedRecipeIds(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		s, ok := ctx.Value(middleware.SessionContextKey).(*session.Session)
		if !ok || s == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]recipebookmark.Bookmark{})
			return
		}

		bookmarks, err := recipebookmark.ByUser(ctx, db, s.User)
		if err != nil {
			logger.Error("failed to list bookmarks", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to list bookmarks"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(bookmarks)
	}
}

func CreateRecipeBookmark(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		s, ok := ctx.Value(middleware.SessionContextKey).(*session.Session)
		if !ok || s == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
			return
		}

		recipeID := r.PathValue("id")
		err := recipebookmark.Create(ctx, db, s.User, recipeID)
		if err != nil {
			logger.Error("failed to bookmark recipe", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to bookmark recipe"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"bookmarked": true})
	}
}

func DeleteRecipeBookmark(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		s, ok := ctx.Value(middleware.SessionContextKey).(*session.Session)
		if !ok || s == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
			return
		}

		recipeID := r.PathValue("id")

		err := recipebookmark.Delete(ctx, db, s.User, recipeID)
		if err != nil {
			logger.Error("failed to unbookmark recipe", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to unbookmark recipe"})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
