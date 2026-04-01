package bookmark

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func Add(ctx context.Context, db *sqlx.DB, userId, recipeId string) error {
	id := uuid.New().String()
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO recipe_bookmarks (bookmark_id, user_id, recipe_id) VALUES (?, ?, ?) ON CONFLICT (user_id, recipe_id) DO NOTHING`,
		id,
		userId,
		recipeId,
	)
	if err != nil {
		return fmt.Errorf("failed to add bookmark: %w", err)
	}
	return nil
}

func Exists(ctx context.Context, db *sqlx.DB, userId, recipeId string) (bool, error) {
	var count int
	err := db.GetContext(
		ctx,
		&count,
		`SELECT COUNT(*) FROM recipe_bookmarks WHERE user_id = ? AND recipe_id = ?`,
		userId,
		recipeId,
	)
	if err != nil {
		return false, fmt.Errorf("failed to check bookmark: %w", err)
	}
	return count > 0, nil
}

func ListByUser(ctx context.Context, db *sqlx.DB, userId string) ([]string, error) {
	var ids []string
	err := db.SelectContext(
		ctx,
		&ids,
		`SELECT recipe_id FROM recipe_bookmarks WHERE user_id = ? ORDER BY created_at DESC`,
		userId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list bookmarks: %w", err)
	}
	if ids == nil {
		ids = []string{}
	}
	return ids, nil
}
