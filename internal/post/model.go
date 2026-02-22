package post

import "time"

type Post struct {
	Id             string     `db:"post_id"         json:"post_id"`
	User           string     `db:"user_id"         json:"user_id"`
	Title          string     `db:"title"           json:"title"`
	Content        string     `db:"content"         json:"content"`
	Public         bool       `db:"public"          json:"public"`
	RevisionId     *string    `db:"revision_id"     json:"revision_id,omitempty"`
	RevisionNumber int        `db:"revision_number" json:"revision_number"`
	CreatedAt      time.Time  `db:"created_at"      json:"created_at"`
	DeletedAt      *time.Time `db:"deleted_at"      json:"deleted_at,omitempty"`
}

type PostWithAuthor struct {
	Post
	Email    string `db:"email"    json:"email"`
	Username string `db:"username" json:"username"`
}
