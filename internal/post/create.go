package post

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CreateRequest struct {
	User    string `json:"user_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Public  bool   `json:"public"`
}

func Create(ctx context.Context, db *sqlx.DB, request CreateRequest) (string, error) {
	postId := uuid.New().String()
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO posts (post_id, user_id, title, content, public) VALUES (?, ?, ?, ?, ?)`,
		postId,
		request.User,
		request.Title,
		request.Content,
		request.Public,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create post: %w", err)
	}

	return postId, nil
}
