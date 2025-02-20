package user

import "time"

type User struct {
	UserId    int64      `db:"user_id"       json:"user_id"`
	Username  string     `db:"username"      json:"username"`
	Email     string     `db:"email"         json:"email"`
	Hash      string     `db:"password_hash" json:"-"`
	Salt      []byte     `db:"salt"          json:"-"`
	Role      string     `db:"role"          json:"role"`
	LastLogin *time.Time `db:"last_login"    json:"last_login,omitempty"`
	CreatedAt time.Time  `db:"created_at"    json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"    json:"updated_at"`
}
