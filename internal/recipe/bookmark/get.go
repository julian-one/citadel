package recipebookmark

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func ByUser(ctx context.Context, db *sqlx.DB, userId string) ([]Bookmark, error) {
	var bookmarks []Bookmark
	err := db.SelectContext(
		ctx,
		&bookmarks,
		`SELECT * FROM recipe_bookmarks WHERE user_id = ? ORDER BY created_at DESC`,
		userId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list bookmarks: %w", err)
	}
	if bookmarks == nil {
		bookmarks = []Bookmark{}
	}
	return bookmarks, nil
}
