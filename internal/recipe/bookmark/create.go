package recipebookmark

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func Create(ctx context.Context, db sqlx.ExecerContext, userId, recipeId string) error {
	id := uuid.New().String()
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO recipe_bookmarks (bookmark_id, user_id, recipe_id) 
			VALUES (?, ?, ?) ON CONFLICT (user_id, recipe_id) DO NOTHING`,
		id,
		userId,
		recipeId,
	)
	if err != nil {
		return fmt.Errorf("failed to add bookmark: %w", err)
	}
	return nil
}
