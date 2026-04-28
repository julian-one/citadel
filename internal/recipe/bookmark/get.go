package recipebookmark

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func ByUser(ctx context.Context, db sqlx.QueryerContext, userId string) ([]Bookmark, error) {
	bookmarks := []Bookmark{}
	err := sqlx.SelectContext(
		ctx,
		db,
		&bookmarks,
		`SELECT * FROM recipe_bookmarks WHERE user_id = ? ORDER BY created_at DESC`,
		userId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list bookmarks: %w", err)
	}

	return bookmarks, nil
}
