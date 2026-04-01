package bookmark

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func Delete(ctx context.Context, db *sqlx.DB, userId, recipeId string) error {
	_, err := db.ExecContext(
		ctx,
		`DELETE FROM recipe_bookmarks WHERE user_id = ? AND recipe_id = ?`,
		userId,
		recipeId,
	)
	if err != nil {
		return fmt.Errorf("failed to delete recipe bookmark: %w", err)
	}
	return nil
}
