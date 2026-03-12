package route

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"citadel/internal/middleware"
	"citadel/internal/recipe"
	"citadel/internal/session"

	"github.com/jmoiron/sqlx"
)

func ListRecipes(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		opts, err := recipe.ParseListOptions(r)
		if err != nil {
			logger.Warn("failed to parse recipe list options", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
			return
		}

		recipes, err := recipe.List(ctx, db, opts)
		if err != nil {
			logger.Error("failed to list recipes", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to list recipes"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(recipes)
	}
}

func CreateRecipe(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		s, ok := ctx.Value(middleware.SessionContextKey).(*session.Session)
		if !ok || s == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
			return
		}

		var req recipe.CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn("failed to decode create recipe request", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		// Set the user from the session to ensure the recipe is associated with the authenticated user
		req.User = s.User

		recipeId, err := recipe.Create(ctx, db, req)
		if err != nil {
			logger.Error("failed to create recipe", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create recipe"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"recipe_id": recipeId})
	}
}

func GetRecipe(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := r.PathValue("id")

		rec, err := recipe.ById(ctx, db, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "Recipe not found"})
				return
			}
			logger.Error("failed to get recipe", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get recipe"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(rec)
	}
}

func UpdateRecipe(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		s, ok := ctx.Value(middleware.SessionContextKey).(*session.Session)
		if !ok || s == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
			return
		}

		id := r.PathValue("id")
		original, err := recipe.ById(ctx, db, id)
		if err != nil {
			logger.Error("failed to fetch recipe for ownership check", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Recipe not found"})
			return
		}

		if s.User != original.User {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "Forbidden"})
			return
		}

		var req recipe.EditableFields
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn("failed to decode update recipe request", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		err = recipe.Update(ctx, db, id, req)
		if err != nil {
			logger.Error("failed to update recipe", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update recipe"})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func DeleteRecipe(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := r.PathValue("id")

		s, ok := ctx.Value(middleware.SessionContextKey).(*session.Session)
		if !ok || s == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
			return
		}

		original, err := recipe.ById(ctx, db, id)
		if err != nil {
			logger.Error("failed to fetch recipe for ownership check", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Recipe not found"})
			return
		}

		if s.User != original.User {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "Forbidden"})
			return
		}

		err = recipe.Delete(ctx, db, id)
		if err != nil {
			logger.Error("failed to delete recipe", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete recipe"})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
