package recipereview

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CreateRequest struct {
	User       string         `json:"user_id"`
	RecipeId   string         `json:"recipe_id"`
	Notes      *string        `json:"notes"`
	Rating     int            `json:"rating"`
	Duration   *time.Duration `json:"duration"`
	Difficulty *int           `json:"difficulty"`
}

func (cr CreateRequest) Validate() error {
	if cr.Rating < 1 || cr.Rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}

	if cr.Difficulty != nil && (*cr.Difficulty < 1 || *cr.Difficulty > 5) {
		return fmt.Errorf("difficulty must be between 1 and 5")
	}

	if cr.Duration != nil && *cr.Duration <= 0 {
		return fmt.Errorf("duration must be greater than 0")
	}

	return nil
}

var ErrDuplicateReview = fmt.Errorf(
	"you can only submit one review per recipe per day",
)

func Create(ctx context.Context, db sqlx.ExtContext, req CreateRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", err
	}

	// Application-level check — gives a clean, user-friendly error.
	var count int
	err := sqlx.GetContext(ctx, db, &count, `
		SELECT COUNT(*) 
		FROM recipe_reviews 
		WHERE user_id = ? AND recipe_id = ? AND date(created_at, 'localtime') = date('now', 'localtime')
	`, req.User, req.RecipeId)
	if err != nil {
		return "", fmt.Errorf("failed to check existing reviews: %w", err)
	}
	if count > 0 {
		return "", ErrDuplicateReview
	}

	id := uuid.New().String()
	_, err = db.ExecContext(
		ctx,
		`INSERT INTO recipe_reviews (review_id, user_id, recipe_id, notes, rating, duration, difficulty) 
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id,
		req.User,
		req.RecipeId,
		req.Notes,
		req.Rating,
		req.Duration,
		req.Difficulty,
	)
	if err != nil {
		// Fallback: the unique index catches anything the app-level check missed.
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return "", ErrDuplicateReview
		}
		return "", fmt.Errorf("failed to create recipe review: %w", err)
	}

	return id, nil
}
