package post

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

func Update(
	ctx context.Context,
	db sqlx.ExtContext,
	postID string,
	edits EditableFields,
) (*Post, error) {
	query := sq.Update("posts").
		Set("updated_at", sq.Expr("datetime('now')")).
		Where(sq.Eq{"post_id": postID}).
		Where(sq.Eq{"deleted_at": nil}).
		Suffix("RETURNING *").
		PlaceholderFormat(sq.Question)

	if edits.Title != nil {
		query = query.Set("title", *edits.Title)
	}
	if edits.Content != nil {
		query = query.Set("content", *edits.Content)
	}
	if edits.Public != nil {
		query = query.Set("public", *edits.Public)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var updated Post
	err = db.QueryRowxContext(ctx, sql, args...).StructScan(&updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}
