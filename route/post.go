package route

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"citadel/internal/post"
	"citadel/internal/session"

	"github.com/jmoiron/sqlx"
)

func ListPosts(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var userID string
		if s, ok := ctx.Value(session.ContextKey).(*session.Session); ok && s != nil {
			userID = s.User
		}

		opts, err := post.ParseListOptions(r, userID)
		if err != nil {
			logger.Warn("failed to parse post list options", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
			return
		}

		items, err := post.List(ctx, db, opts)
		if err != nil {
			logger.Error("failed to list posts", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to list posts"})
			return
		}

		total, err := post.Count(ctx, db, opts)
		if err != nil {
			logger.Error("failed to count posts", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to list posts"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct {
			Items []post.PostWithAuthor `json:"items"`
			Total int                   `json:"total"`
		}{
			Items: items,
			Total: total,
		})
	}
}

func CreatePost(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req post.CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn("failed to decode create post request", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		postID, err := post.Create(r.Context(), db, req)
		if err != nil {
			logger.Error("failed to create post", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create post"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"post_id": postID})
	}
}

func GetPost(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := r.PathValue("id")

		p, err := post.ByIdWithAuthor(ctx, db, id)
		if err != nil {
			logger.Error("failed to get post", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get post"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(p)
	}
}

func UpdatePost(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		s, ok := ctx.Value(session.ContextKey).(*session.Session)
		if !ok || s == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
			return
		}

		id := r.PathValue("id")
		original, err := post.ById(ctx, db, id)
		if err != nil {
			logger.Error("failed to fetch post for ownership check", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Post not found"})
			return
		}

		if s.User != original.User {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "Forbidden"})
			return
		}

		var req post.EditableFields
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			logger.Warn("failed to decode update post request", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		_, err = post.Update(ctx, db, id, req)
		if err != nil {
			logger.Error("failed to update post", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update post"})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func DeletePost(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := r.PathValue("id")

		s, ok := ctx.Value(session.ContextKey).(*session.Session)
		if !ok || s == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
			return
		}

		original, err := post.ById(ctx, db, id)
		if err != nil {
			logger.Error("failed to fetch post for ownership check", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Post not found"})
			return
		}

		if s.User != original.User {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "Forbidden"})
			return
		}

		err = post.Delete(ctx, db, id)
		if err != nil {
			logger.Error("failed to delete post", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete post"})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
