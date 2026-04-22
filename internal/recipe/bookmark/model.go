package recipebookmark

import (
	"time"
)

type Bookmark struct {
	Id        string    `db:"bookmark_id" json:"bookmark_id"`
	User      string    `db:"user_id"     json:"user_id"`
	RecipeId  string    `db:"recipe_id"   json:"recipe_id"`
	CreatedAt time.Time `db:"created_at"  json:"created_at"`
}
