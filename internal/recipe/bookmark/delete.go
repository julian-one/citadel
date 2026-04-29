package recipebookmark

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func Delete(ctx context.Context, db sqlx.ExecerContext, userID, recipeID string) error {
	_, err := db.ExecContext(
		ctx,
		`DELETE FROM recipe_bookmarks WHERE user_id = ? AND recipe_id = ?`,
		userID,
		recipeID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete recipe bookmark: %w", err)
	}
	return nil
}
