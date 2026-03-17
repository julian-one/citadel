package user

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func UpdatePassword(
	ctx context.Context,
	db *sqlx.DB,
	userId string,
	newPassword string,
) error {
	h, s, err := Hash(newPassword, nil)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	_, err = db.ExecContext(
		ctx,
		`UPDATE users SET password_hash = ?, salt = ?, updated_at = datetime('now') WHERE user_id = ?`,
		h,
		s,
		userId,
	)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}
