package recipe

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func Update(
	ctx context.Context,
	db *sqlx.DB,
	recipeId string,
	edits EditableFields,
) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := sq.Update("recipes").
		Set("updated_at", sq.Expr("datetime('now')")).
		Where(sq.Eq{"recipe_id": recipeId}).
		Where(sq.Eq{"deleted_at": nil}).
		PlaceholderFormat(sq.Question)

	hasRecipeUpdates := false
	if edits.Title != nil {
		query = query.Set("title", *edits.Title)
		hasRecipeUpdates = true
	}
	if edits.Description != nil {
		query = query.Set("description", *edits.Description)
		hasRecipeUpdates = true
	}
	if edits.PhotoUrl != nil {
		query = query.Set("photo_url", *edits.PhotoUrl)
		hasRecipeUpdates = true
	}
	if edits.SourceUrl != nil {
		query = query.Set("source_url", *edits.SourceUrl)
		hasRecipeUpdates = true
	}
	if edits.CookTime != nil {
		query = query.Set("cook_time", *edits.CookTime)
		hasRecipeUpdates = true
	}
	if edits.Serves != nil {
		query = query.Set("serves", *edits.Serves)
		hasRecipeUpdates = true
	}
	if edits.Cuisine != nil {
		query = query.Set("cuisine", *edits.Cuisine)
		hasRecipeUpdates = true
	}
	if edits.Category != nil {
		query = query.Set("category", *edits.Category)
		hasRecipeUpdates = true
	}

	if hasRecipeUpdates {
		sql, args, err := query.ToSql()
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to update recipe: %w", err)
		}
	}

	if edits.Ingredients != nil {
		_, err = tx.ExecContext(ctx, `DELETE FROM ingredients WHERE recipe_id = ?`, recipeId)
		if err != nil {
			return fmt.Errorf("failed to delete old ingredients: %w", err)
		}

		for _, ing := range *edits.Ingredients {
			iid := uuid.New().String()
			_, err = tx.ExecContext(
				ctx,
				`INSERT INTO ingredients (ingredient_id, recipe_id, amount, unit, item) VALUES (?, ?, ?, ?, ?)`,
				iid,
				recipeId,
				ing.Amount,
				ing.Unit,
				ing.Item,
			)
			if err != nil {
				return fmt.Errorf("failed to insert new ingredient: %w", err)
			}
		}
	}

	if edits.Instructions != nil {
		_, err = tx.ExecContext(
			ctx,
			`DELETE FROM instructions WHERE recipe_id = ?`,
			recipeId,
		)
		if err != nil {
			return fmt.Errorf("failed to delete old instructions: %w", err)
		}

		for i, instr := range *edits.Instructions {
			inid := uuid.New().String()
			_, err = tx.ExecContext(
				ctx,
				`INSERT INTO instructions (instruction_id, recipe_id, step_number, instruction) VALUES (?, ?, ?, ?)`,
				inid,
				recipeId,
				i+1,
				instr,
			)
			if err != nil {
				return fmt.Errorf("failed to insert new instruction: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
