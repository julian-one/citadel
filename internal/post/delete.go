package post

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func Delete(ctx context.Context, db *sqlx.DB, postId string) error {
	_, err := db.ExecContext(
		ctx,
		`UPDATE posts SET deleted_at = datetime('now') WHERE post_id = ?`,
		postId,
	)
	if err != nil {
		return fmt.Errorf("failed to mark post as deleted: %w", err)
	}
	return nil
}
