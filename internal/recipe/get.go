package recipe

import (
	"context"
	"encoding/json"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

func ById(ctx context.Context, db *sqlx.DB, recipeId string) (*Recipe, error) {
	var row recipeRow

	queryBuilder := sq.Select("r.*").
		Column("(SELECT json_group_array(json_object('amount', amount, 'unit', unit, 'item', item)) FROM ingredients WHERE recipe_id = r.recipe_id) AS ingredients_json").
		Column("(SELECT json_group_array(instruction) FROM instructions WHERE recipe_id = r.recipe_id ORDER BY step_number ASC) AS instructions_json").
		From("recipes r").
		Where(sq.Eq{"r.recipe_id": recipeId}).
		Where(sq.Eq{"r.deleted_at": nil}).
		PlaceholderFormat(sq.Question)

	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	err = db.GetContext(ctx, &row, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe by id: %w", err)
	}

	if len(row.RawIngredients) > 0 {
		if err := json.Unmarshal(row.RawIngredients, &row.Recipe.Ingredients); err != nil {
			return nil, fmt.Errorf("failed to unmarshal ingredients: %w", err)
		}
	}

	if len(row.RawInstructions) > 0 {
		if err := json.Unmarshal(row.RawInstructions, &row.Recipe.Instructions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal instructions: %w", err)
		}
	}

	return &row.Recipe, nil
}
