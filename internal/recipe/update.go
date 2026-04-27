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
	recipeID string,
	edits EditableFields,
) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := sq.Update("recipes").
		Set("updated_at", sq.Expr("datetime('now')")).
		Where(sq.Eq{"recipe_id": recipeID}).
		Where(sq.Eq{"deleted_at": nil}).
		PlaceholderFormat(sq.Question)

	hasRecipeUpdates := false
	if edits.Title != nil {
		query = query.Set("title", *edits.Title)
		hasRecipeUpdates = true
	}
	if edits.Description != nil {
		if *edits.Description == "" {
			query = query.Set("description", nil)
		} else {
			query = query.Set("description", *edits.Description)
		}
		hasRecipeUpdates = true
	}
	if edits.PhotoURL != nil {
		if *edits.PhotoURL == "" {
			query = query.Set("photo_url", nil)
		} else {
			query = query.Set("photo_url", *edits.PhotoURL)
		}
		hasRecipeUpdates = true
	}
	if edits.SourceType != nil {
		if *edits.SourceType == "" {
			query = query.Set("source_type", nil)
		} else {
			query = query.Set("source_type", *edits.SourceType)
		}
		hasRecipeUpdates = true
	}
	if edits.Source != nil {
		if *edits.Source == "" {
			query = query.Set("source", nil)
		} else {
			query = query.Set("source", *edits.Source)
		}
		hasRecipeUpdates = true
	}
	if edits.PrepTime != nil {
		if *edits.PrepTime == 0 {
			query = query.Set("prep_time", nil)
		} else {
			query = query.Set("prep_time", *edits.PrepTime)
		}
		hasRecipeUpdates = true
	}
	if edits.CookTime != nil {
		if *edits.CookTime == 0 {
			query = query.Set("cook_time", nil)
		} else {
			query = query.Set("cook_time", *edits.CookTime)
		}
		hasRecipeUpdates = true
	}
	if edits.Serves != nil {
		if *edits.Serves == 0 {
			query = query.Set("serves", nil)
		} else {
			query = query.Set("serves", *edits.Serves)
		}
		hasRecipeUpdates = true
	}
	if edits.Cuisine != nil {
		if *edits.Cuisine == "" {
			query = query.Set("cuisine", nil)
		} else {
			query = query.Set("cuisine", *edits.Cuisine)
		}
		hasRecipeUpdates = true
	}
	if edits.Category != nil {
		if *edits.Category == "" {
			query = query.Set("category", nil)
		} else {
			query = query.Set("category", *edits.Category)
		}
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

	if edits.Components != nil {
		// Cascade delete removes all components, ingredients, and instructions
		_, err = tx.ExecContext(
			ctx,
			`DELETE FROM recipe_components WHERE recipe_id = ?`,
			recipeID,
		)
		if err != nil {
			return fmt.Errorf("failed to delete old components: %w", err)
		}

		for i, comp := range *edits.Components {
			cid := uuid.New().String()
			_, err = tx.ExecContext(
				ctx,
				`INSERT INTO recipe_components (component_id, recipe_id, name, position) VALUES (?, ?, ?, ?)`,
				cid,
				recipeID,
				comp.Name,
				i,
			)
			if err != nil {
				return fmt.Errorf("failed to insert component: %w", err)
			}

			for _, ing := range comp.Ingredients {
				iid := uuid.New().String()
				_, err = tx.ExecContext(
					ctx,
					`INSERT INTO ingredients (ingredient_id, component_id, amount, unit, item) VALUES (?, ?, ?, ?, ?)`,
					iid,
					cid,
					ing.Amount,
					ing.Unit,
					ing.Item,
				)
				if err != nil {
					return fmt.Errorf("failed to insert ingredient: %w", err)
				}
			}

			for j, instr := range comp.Instructions {
				inid := uuid.New().String()
				_, err = tx.ExecContext(
					ctx,
					`INSERT INTO instructions (instruction_id, component_id, step_number, instruction) VALUES (?, ?, ?, ?)`,
					inid,
					cid,
					j+1,
					instr,
				)
				if err != nil {
					return fmt.Errorf("failed to insert instruction: %w", err)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
