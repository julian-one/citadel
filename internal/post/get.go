package post

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

func ById(ctx context.Context, db *sqlx.DB, postId string) (*PostWithAuthor, error) {
	query := sq.Select("p.*, u.email, u.username").From("posts p").
		InnerJoin("users u ON (u.user_id = p.user_id)").
		Where(sq.Expr("COALESCE(p.revision_id, p.post_id) = ?", postId)).
		Where("p.deleted_at IS NULL").
		OrderBy("p.revision_number DESC").
		Limit(1)

	sql, args, err := query.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var p PostWithAuthor
	err = db.GetContext(ctx, &p, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get post by id: %w", err)
	}
	return &p, nil
}
