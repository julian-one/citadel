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
	Search     string // searches username and email
	Role       *Role  // filter by role (nil = no filter)
	OrderBy    []database.Order
	Pagination database.Pagination
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
		parsed, err := database.ParseOrder(orderBy)
		if err != nil {
			return opts, err
		}
		for _, o := range parsed {
			if !isValidColumn(o.Column) {
				return opts, fmt.Errorf("invalid column name: %s", o.Column)
			}
		}
		opts.OrderBy = parsed
	}

	pagination, err := database.ParsePagination(query.Get("limit"), query.Get("offset"))
	if err != nil {
		return opts, err
	}
	opts.Pagination = pagination

	return opts, nil
}

func applyUserFilters(q sq.SelectBuilder, opts ListOptions) sq.SelectBuilder {
	if opts.Search != "" {
		searchPattern := "%" + strings.ToLower(opts.Search) + "%"
		q = q.Where(sq.Or{
			sq.Expr("LOWER(username) LIKE ?", searchPattern),
			sq.Expr("LOWER(email) LIKE ?", searchPattern),
		})
	}
	if opts.Role != nil {
		q = q.Where(sq.Eq{"role": *opts.Role})
	}
	return q
}

func Count(ctx context.Context, db *sqlx.DB, opts ListOptions) (int, error) {
	q := applyUserFilters(sq.Select("COUNT(*)").From("users"), opts)

	sql, args, err := q.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build count query: %w", err)
	}

	var total int
	if err = db.QueryRowContext(ctx, sql, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return total, nil
}

func List(ctx context.Context, db *sqlx.DB, opts ListOptions) ([]User, error) {
	if opts.Role != nil && !opts.Role.Valid() {
		return nil, fmt.Errorf("invalid role: %s", *opts.Role)
	}

	query := applyUserFilters(sq.Select("*").From("users"), opts)

	if len(opts.OrderBy) > 0 {
		for _, o := range opts.OrderBy {
			if !isValidColumn(o.Column) {
				return nil, fmt.Errorf("invalid column name: %s", o.Column)
			}
			query = query.OrderBy(fmt.Sprintf("%s %s", o.Column, o.Direction))
		}
	} else {
		query = query.OrderBy("created_at DESC")
	}

	query = query.Limit(uint64(opts.Pagination.Limit)).Offset(uint64(opts.Pagination.Offset))

	sql, args, err := query.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	users := []User{}
	err = db.SelectContext(ctx, &users, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	if users == nil {
		users = []User{}
	}

	return users, nil
}
