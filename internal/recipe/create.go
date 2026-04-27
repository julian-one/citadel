package recipe

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ComponentRequest struct {
	Name         *string      `json:"name"`
	Ingredients  []Ingredient `json:"ingredients"`
	Instructions []string     `json:"instructions"`
}

type CreateRequest struct {
	User        string             `json:"user_id"`
	Title       string             `json:"title"`
	Description *string            `json:"description"`
	Components  []ComponentRequest `json:"components"`
	PrepTime    *time.Duration     `json:"prep_time"`
	CookTime    *time.Duration     `json:"cook_time"`
	Serves      *uint32            `json:"serves"`
	Cuisine     *Cuisine           `json:"cuisine"`
	Category    *Category          `json:"category"`
	PhotoURL    *string            `json:"photo_url"`
	SourceType  *SourceType        `json:"source_type"`
	Source      *string            `json:"source"`
}

func Create(ctx context.Context, db *sqlx.DB, request CreateRequest) (string, error) {
	rid := uuid.New().String()

	tx, err := db.Beginx()
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO recipes (recipe_id, user_id, title, description, photo_url, source_type, source, prep_time, cook_time, serves, cuisine, category) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rid,
		request.User,
		request.Title,
		request.Description,
		request.PhotoURL,
		request.SourceType,
		request.Source,
		request.PrepTime,
		request.CookTime,
		request.Serves,
		request.Cuisine,
		request.Category,
	)
	if err != nil {
		return "", fmt.Errorf("failed to insert recipe: %w", err)
	}

	for i, comp := range request.Components {
		cid := uuid.New().String()
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO recipe_components (component_id, recipe_id, name, position) VALUES (?, ?, ?, ?)`,
			cid,
			rid,
			comp.Name,
			i,
		)
		if err != nil {
			return "", fmt.Errorf("failed to insert component: %w", err)
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
				return "", fmt.Errorf("failed to insert ingredient: %w", err)
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
				return "", fmt.Errorf("failed to insert instruction: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return rid, nil
}
