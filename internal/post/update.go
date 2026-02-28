package post

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

func Update(ctx context.Context, db *sqlx.DB, postId string, edits EditableFields) (*Post, error) {
	query := sq.Update("posts").
		Set("title", edits.Title).
		Set("content", edits.Content).
		Set("public", edits.Public).
		Set("updated_at", sq.Expr("datetime('now')")).
		Where(sq.Eq{"post_id": postId}).
		Where(sq.Eq{"deleted_at": nil}).
		Suffix("RETURNING *").
		PlaceholderFormat(sq.Question)

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
