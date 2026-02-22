package user

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"citadel/internal/database"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

// isValidColumn validates column names against a whitelist to prevent SQL injection
func isValidColumn(column string) bool {
	validColumns := map[string]bool{
		"user_id":    true,
		"username":   true,
		"email":      true,
		"role":       true,
		"created_at": true,
		"updated_at": true,
	}
	return validColumns[column]
}

// ListOptions contains options for listing users
type ListOptions struct {
	Search  string // searches username and email
	Role    *Role  // filter by role (nil = no filter)
	OrderBy []database.OrderBy
}

func ParseListOptions(r *http.Request) (ListOptions, error) {
	query := r.URL.Query()

	var opts ListOptions

	// Parse search parameter
	if search := query.Get("search"); search != "" {
		opts.Search = search
	}

	// Parse role filter
	if roleStr := query.Get("role"); roleStr != "" {
		role := Role(roleStr)
		if !role.Valid() {
			return opts, fmt.Errorf("invalid role: %s (must be 'admin' or 'user')", roleStr)
		}
		opts.Role = &role
	}

	if orderBy := query.Get("order_by"); orderBy != "" {
		parsed, err := database.ParseOrderBy(orderBy)
		if err != nil {
			return opts, err
		}
		opts.OrderBy = parsed
	}

	return opts, nil
}

func List(ctx context.Context, db *sqlx.DB, opts ListOptions) ([]User, error) {
	query := sq.Select("*").From("users")

	// Apply search filter
	if opts.Search != "" {
		searchPattern := "%" + strings.ToLower(opts.Search) + "%"
		query = query.Where(sq.Or{
			sq.Expr("LOWER(username) LIKE ?", searchPattern),
			sq.Expr("LOWER(email) LIKE ?", searchPattern),
		})
	}

	// Apply role filter
	if opts.Role != nil {
		if !opts.Role.Valid() {
			return nil, fmt.Errorf("invalid role: %s", *opts.Role)
		}
		query = query.Where(sq.Eq{"role": *opts.Role})
	}

	// Apply ordering
	if len(opts.OrderBy) > 0 {
		for _, o := range opts.OrderBy {
			// Validate order direction
			if !o.Order.Valid() {
				return nil, fmt.Errorf("invalid order direction: %s", o.Order)
			}

			// Validate column name (whitelist approach for security)
			if !isValidColumn(o.Column) {
				return nil, fmt.Errorf("invalid column name: %s", o.Column)
			}

			query = query.OrderBy(fmt.Sprintf("%s %s", o.Column, o.Order))
		}
	} else {
		// Default ordering if none specified
		query = query.OrderBy("created_at DESC")
	}

	sql, args, err := query.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var users []User
	err = db.SelectContext(ctx, &users, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}
