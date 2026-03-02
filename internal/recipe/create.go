package recipe

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CreateRequest struct {
	User         string         `json:"user_id"`
	Title        string         `json:"title"`
	Description  *string        `json:"description"`
	Ingredients  []Ingredient   `json:"ingredients"`
	Instructions []string       `json:"instructions"`
	CookTime     *time.Duration `json:"cook_time"`
	Serves       *uint32        `json:"serves"`
	Cuisine      *Cuisine       `json:"cuisine"`
	Category     *Category      `json:"category"`
	PhotoUrl     *string        `json:"photo_url"`
	SourceUrl    *string        `json:"source_url"`
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
		`INSERT INTO recipes (recipe_id, user_id, title, description, photo_url, source_url, cook_time, serves, cuisine, category) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rid,
		request.User,
		request.Title,
		request.Description,
		request.PhotoUrl,
		request.SourceUrl,
		request.CookTime,
		request.Serves,
		request.Cuisine,
		request.Category,
	)
	if err != nil {
		return "", fmt.Errorf("failed to insert recipe: %w", err)
	}

	for _, ing := range request.Ingredients {
		iid := uuid.New().String()
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO ingredients (ingredient_id, recipe_id, amount, unit, item) VALUES (?, ?, ?, ?, ?)`,
			iid,
			rid,
			ing.Amount,
			ing.Unit,
			ing.Item,
		)
		if err != nil {
			return "", fmt.Errorf("failed to insert ingredient: %w", err)
		}
	}

	for i, instr := range request.Instructions {
		inid := uuid.New().String()
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO instructions (instruction_id, recipe_id, step_number, instruction) VALUES (?, ?, ?, ?)`,
			inid,
			rid,
			i+1,
			instr,
		)
		if err != nil {
			return "", fmt.Errorf("failed to insert instruction: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return rid, nil
}
