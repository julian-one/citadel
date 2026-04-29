package post

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

func ByID(ctx context.Context, db sqlx.QueryerContext, postID string) (*Post, error) {
	query := sq.Select("*").From("posts").
		Where("post_id = ?", postID).
		Where("deleted_at IS NULL")

	sql, args, err := query.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var p Post
	err = sqlx.GetContext(ctx, db, &p, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get post by id: %w", err)
	}
	return &p, nil
}

func ByIDWithAuthor(
	ctx context.Context,
	db sqlx.QueryerContext,
	postID string,
) (*PostWithAuthor, error) {
	query := sq.Select("p.*, u.email, u.username").From("posts p").
		InnerJoin("users u ON (u.user_id = p.user_id)").
		Where("p.post_id = ?", postID).
		Where("p.deleted_at IS NULL")

	sql, args, err := query.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var p PostWithAuthor
	err = sqlx.GetContext(ctx, db, &p, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get post with author by id: %w", err)
	}
	return &p, nil
}
