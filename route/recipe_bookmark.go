package route

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"citadel/internal/middleware"
	"citadel/internal/recipe/bookmark"
	"citadel/internal/session"

	"github.com/jmoiron/sqlx"
)

func BookmarkRecipe(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		s, ok := ctx.Value(middleware.SessionContextKey).(*session.Session)
		if !ok || s == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
			return
		}

		recipeId := r.PathValue("id")

		err := bookmark.Add(ctx, db, s.User, recipeId)
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

		recipeId := r.PathValue("id")

		err := bookmark.Delete(ctx, db, s.User, recipeId)
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

func GetBookmarkStatus(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		s, ok := ctx.Value(middleware.SessionContextKey).(*session.Session)
		if !ok || s == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
			return
		}

		recipeId := r.PathValue("id")

		exists, err := bookmark.Exists(ctx, db, s.User, recipeId)
		if err != nil {
			logger.Error("failed to check bookmark status", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to check bookmark status"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"bookmarked": exists})
	}
}

func ListBookmarkedRecipeIds(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		s, ok := ctx.Value(middleware.SessionContextKey).(*session.Session)
		if !ok || s == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]string{})
			return
		}

		ids, err := bookmark.ListByUser(ctx, db, s.User)
		if err != nil {
			logger.Error("failed to list bookmarked recipe ids", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to list bookmarks"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ids)
	}
}
