package post

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

func ListRevisions(ctx context.Context, db *sqlx.DB, postId string) ([]Post, error) {
	query := sq.Select("*").From("posts").
		Where(sq.Expr("COALESCE(revision_id, post_id) = ?", postId)).
		Where("deleted_at IS NULL").
		OrderBy("revision_number ASC")

	sql, args, err := query.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var posts []Post
	err = db.SelectContext(ctx, &posts, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list revisions: %w", err)
	}

	return posts, nil
}
