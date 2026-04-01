package log

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func Create(ctx context.Context, db *sqlx.DB, req CreateRequest) (string, error) {
	id := uuid.New().String()

	if req.Rating != nil {
		r := *req.Rating
		if r < 1.0 || r > 5.0 {
			return "", fmt.Errorf("rating must be between 1.0 and 5.0")
		}
	}

	if req.Intensity != nil {
		i := *req.Intensity
		if i < 1 || i > 3 {
			return "", fmt.Errorf("intensity must be 1, 2, or 3")
		}
	}

	_, err := db.ExecContext(
		ctx,
		`INSERT INTO recipe_logs (log_id, user_id, recipe_id, notes, rating, duration, intensity) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id,
		req.User,
		req.RecipeId,
		req.Notes,
		req.Rating,
		req.Duration,
		req.Intensity,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create recipe log: %w", err)
	}

	return id, nil
}

func ListByRecipe(
	ctx context.Context,
	db *sqlx.DB,
	userId, recipeId string,
) ([]RecipeLog, error) {
	logs := make([]RecipeLog, 0)
	err := db.SelectContext(
		ctx,
		&logs,
		`SELECT * FROM recipe_logs WHERE user_id = ? AND recipe_id = ? ORDER BY created_at DESC`,
		userId,
		recipeId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list recipe logs: %w", err)
	}
	return logs, nil
}

func Delete(ctx context.Context, db *sqlx.DB, userId, logId string) error {
	result, err := db.ExecContext(
		ctx,
		`DELETE FROM recipe_logs WHERE log_id = ? AND user_id = ?`,
		logId,
		userId,
	)
	if err != nil {
		return fmt.Errorf("failed to delete recipe log: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to delete recipe log: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("recipe log not found or not owned by user")
	}

	return nil
}
