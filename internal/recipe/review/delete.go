package recipereview

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func Delete(ctx context.Context, db *sqlx.DB, userId, reviewId string) error {
	_, err := db.ExecContext(
		ctx,
		`DELETE FROM recipe_reviews WHERE review_id = ? AND user_id = ?`,
		reviewId,
		userId,
	)
	if err != nil {
		return fmt.Errorf("failed to delete recipe review: %w", err)
	}

	return nil
}
