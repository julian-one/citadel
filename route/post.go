package route

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"citadel/internal/middleware"
	"citadel/internal/post"
	"citadel/internal/session"

	"github.com/jmoiron/sqlx"
)

func ListPosts(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("ListPosts called")
		ctx := r.Context()

		var userID string
		if s, ok := ctx.Value(middleware.SessionContextKey).(*session.Session); ok && s != nil {
			userID = s.User
		}

		opts, err := post.ParseListOptions(r, userID)
		if err != nil {
			logger.Error("Failed to parse post list options", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
			return
		}

		posts, err := post.List(ctx, db, opts)
		if err != nil {
			logger.Error("Failed to list posts", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to list posts"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(posts)
	}
}

func CreatePost(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("CreatePost called")

		var req post.CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("Failed to decode create post request", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		postId, err := post.Create(r.Context(), db, req)
		if err != nil {
			logger.Error("Failed to create post", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create post"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"post_id": postId})
	}
}

func GetPost(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("GetPost called")

		ctx := r.Context()
		id := r.PathValue("id")

		p, err := post.ById(ctx, db, id)
		if err != nil {
			logger.Error("Failed to get post", "error", err)
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
		logger.Info("UpdatePost called")

		id := r.PathValue("id")

		var req post.UpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("Failed to decode update post request", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}
		req.PostId = id

		newId, err := post.Update(r.Context(), db, req)
		if err != nil {
			logger.Error("Failed to update post", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update post"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"post_id": newId})
	}
}

func ListPostRevisions(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("ListPostRevisions called")

		id := r.PathValue("id")

		revisions, err := post.ListRevisions(r.Context(), db, id)
		if err != nil {
			logger.Error("Failed to list post revisions", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to list revisions"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(revisions)
	}
}

func DeletePost(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("DeletePost called")

		ctx := r.Context()
		id := r.PathValue("id")
		err := post.Delete(ctx, db, id)
		if err != nil {
			logger.Error("Failed to delete post", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete post"})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
