package log

import (
	"time"
)

type RecipeLog struct {
	Id        string         `db:"log_id"     json:"log_id"`
	User      string         `db:"user_id"    json:"user_id"`
	RecipeId  string         `db:"recipe_id"  json:"recipe_id"`
	Notes     *string        `db:"notes"      json:"notes"`
	Rating    *float64       `db:"rating"     json:"rating"`
	Duration  *time.Duration `db:"duration"   json:"duration"`
	Intensity *int           `db:"intensity"  json:"intensity"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
}

type CreateRequest struct {
	User      string         `json:"user_id"`
	RecipeId  string         `json:"recipe_id"`
	Notes     *string        `json:"notes"`
	Rating    *float64       `json:"rating"`
	Duration  *time.Duration `json:"duration"`
	Intensity *int           `json:"intensity"`
}
