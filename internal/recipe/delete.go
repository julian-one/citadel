package recipe

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func Delete(ctx context.Context, db sqlx.ExecerContext, recipeID string) error {
	_, err := db.ExecContext(
		ctx,
		`UPDATE recipes SET deleted_at = datetime('now') WHERE recipe_id = ?`,
		recipeID,
	)
	if err != nil {
		return fmt.Errorf("failed to mark recipe as deleted: %w", err)
	}
	return nil
}
