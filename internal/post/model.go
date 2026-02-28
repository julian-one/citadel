package post

import "time"

type Post struct {
	Id        string     `db:"post_id"    json:"post_id"`
	User      string     `db:"user_id"    json:"user_id"`
	Title     string     `db:"title"      json:"title"`
	Content   string     `db:"content"    json:"content"`
	Public    bool       `db:"public"     json:"public"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

type PostWithAuthor struct {
	Post
	Email    string `db:"email"    json:"email"`
	Username string `db:"username" json:"username"`
}

type EditableFields struct {
	Title   string `db:"title"   json:"title"`
	Content string `db:"content" json:"content"`
	Public  bool   `db:"public"  json:"public"`
}
