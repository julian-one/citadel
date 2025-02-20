package user

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/scrypt"
)

type CreateRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Create(ctx context.Context, db *sqlx.DB, request CreateRequest) (int64, error) {
	h, s, err := hash(request.Password, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	result, err := db.ExecContext(
		ctx,
		`INSERT INTO users (username, email, password_hash, salt) VALUES (?, ?, ?, ?)`,
		request.Username,
		request.Email,
		h,
		s,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}
	uid, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}
	return uid, nil
}

func Verify(password string, storedHash string, salt []byte) (bool, error) {
	computed, _, err := hash(password, salt)
	if err != nil {
		return false, fmt.Errorf("failed to compute hash: %w", err)
	}
	match := subtle.ConstantTimeCompare([]byte(computed), []byte(storedHash)) == 1
	return match, nil
}

// hash uses scrypt with OWASP 2024 recommended parameters.
// N=32768 (2^15), r=8, p=1, keyLen=32
// See: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html
func hash(password string, salt []byte) (string, []byte, error) {
	if salt == nil {
		salt = make([]byte, 32)
		if _, err := rand.Read(salt); err != nil {
			return "", nil, fmt.Errorf("failed to generate salt: %w", err)
		}
	}

	hash, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(hash), salt, nil
}
