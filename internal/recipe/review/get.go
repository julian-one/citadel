package recipereview

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type ReviewResponse struct {
	RecipeReview
	Username string `db:"username" json:"username"`
}

func ByRecipe(
	ctx context.Context,
	db sqlx.QueryerContext,
	recipeId string,
) ([]ReviewResponse, error) {
	reviews := make([]ReviewResponse, 0)
	err := sqlx.SelectContext(
		ctx,
		db,
		&reviews,
		`SELECT r.*, u.username 
		 FROM recipe_reviews r 
		 JOIN users u ON r.user_id = u.user_id 
		 WHERE r.recipe_id = ? 
		 ORDER BY r.created_at DESC`,
		recipeId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list recipe reviews: %w", err)
	}
	return reviews, nil
}
