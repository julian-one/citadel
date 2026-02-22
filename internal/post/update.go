package post

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UpdateRequest struct {
	PostId  string `json:"post_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Public  bool   `json:"public"`
}

func Update(ctx context.Context, db *sqlx.DB, request UpdateRequest) (string, error) {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get the max revision number for this post chain
	var maxRev int
	revQuery, revArgs, err := sq.Select("MAX(revision_number)").From("posts").
		Where(sq.Expr("COALESCE(revision_id, post_id) = ?", request.PostId)).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return "", fmt.Errorf("failed to build revision query: %w", err)
	}
	err = tx.GetContext(ctx, &maxRev, revQuery, revArgs...)
	if err != nil {
		return "", fmt.Errorf("failed to get max revision: %w", err)
	}

	// Get the original post's user_id and public flag
	var original struct {
		User   string `db:"user_id"`
		Public bool   `db:"public"`
	}
	origQuery, origArgs, err := sq.Select("user_id", "public").From("posts").
		Where(sq.Eq{"post_id": request.PostId}).
		Where("deleted_at IS NULL").
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return "", fmt.Errorf("failed to build original query: %w", err)
	}
	err = tx.GetContext(ctx, &original, origQuery, origArgs...)
	if err != nil {
		return "", fmt.Errorf("failed to get original post: %w", err)
	}

	// Insert new revision
	newId := uuid.New().String()
	insertQuery, insertArgs, err := sq.Insert("posts").
		Columns("post_id", "user_id", "title", "content", "public", "revision_id", "revision_number").
		Values(newId, original.User, request.Title, request.Content, request.Public, request.PostId, maxRev+1).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return "", fmt.Errorf("failed to build insert query: %w", err)
	}
	_, err = tx.ExecContext(ctx, insertQuery, insertArgs...)
	if err != nil {
		return "", fmt.Errorf("failed to insert revision: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newId, nil
}
