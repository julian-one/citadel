package recipe

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func ByID(ctx context.Context, db *sqlx.DB, recipeID string) (*Recipe, error) {
	var r Recipe

	err := db.GetContext(
		ctx,
		&r,
		`SELECT * FROM recipes WHERE recipe_id = ? AND deleted_at IS NULL`,
		recipeID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe by id: %w", err)
	}

	components, err := LoadComponents(ctx, db, []string{r.ID})
	if err != nil {
		return nil, err
	}

	r.Components = components
	if r.Components == nil {
		r.Components = []Component{}
	}

	return &r, nil
}
