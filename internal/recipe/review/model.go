package recipereview

import (
	"time"
)

type RecipeReview struct {
	Id         string         `db:"review_id"  json:"review_id"`
	User       string         `db:"user_id"    json:"user_id"`
	RecipeId   string         `db:"recipe_id"  json:"recipe_id"`
	Notes      *string        `db:"notes"      json:"notes"`
	Rating     int            `db:"rating"     json:"rating"`
	Duration   *time.Duration `db:"duration"   json:"duration"`
	Difficulty *int           `db:"difficulty" json:"difficulty"`
	CreatedAt  time.Time      `db:"created_at" json:"created_at"`
}
