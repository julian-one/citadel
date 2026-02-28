package post

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"citadel/internal/database"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

// isValidColumn validates column names against a whitelist to prevent SQL injection
func isValidColumn(column string) bool {
	validColumns := map[string]bool{
		"post_id":    true,
		"user_id":    true,
		"title":      true,
		"body":       true,
		"public":     true,
		"created_at": true,
	}
	return validColumns[column]
}

type ListOptions struct {
	Search  string
	User    string
	Public  *bool
	OrderBy []database.OrderBy
}

func ParseListOptions(r *http.Request, user string) (ListOptions, error) {
	query := r.URL.Query()

	var opts ListOptions

	if search := query.Get("search"); search != "" {
		opts.Search = search
	}
	opts.User = user

	if public := query.Get("public"); public != "" {
		b, err := strconv.ParseBool(public)
		if err != nil {
			return opts, fmt.Errorf("invalid public value: %s (must be 'true' or 'false')", public)
		}
		opts.Public = &b
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

func List(ctx context.Context, db *sqlx.DB, opts ListOptions) ([]PostWithAuthor, error) {
	query := sq.Select("p.*, u.email, u.username").
		From("posts p").
		InnerJoin("users u ON (u.user_id = p.user_id)").
		Where("p.deleted_at IS NULL")

	if opts.Search != "" {
		searchPattern := "%" + strings.ToLower(opts.Search) + "%"
		query = query.Where(sq.Expr("LOWER(p.title) LIKE ?", searchPattern))
	}
	if opts.User != "" && opts.Public == nil {
		// Authenticated user with no explicit public filter:
		// show their own posts (public + private) and others' public posts
		query = query.Where(sq.Or{
			sq.Eq{"p.user_id": opts.User},
			sq.Eq{"p.public": true},
		})
	} else {
		if opts.User != "" {
			query = query.Where(sq.Eq{"p.user_id": opts.User})
		}
		if opts.Public != nil {
			query = query.Where(sq.Eq{"p.public": *opts.Public})
		}
	}

	if len(opts.OrderBy) > 0 {
		for _, o := range opts.OrderBy {
			if !o.Order.Valid() {
				return nil, fmt.Errorf("invalid order direction: %s", o.Order)
			}

			if !isValidColumn(o.Column) {
				return nil, fmt.Errorf("invalid column name: %s", o.Column)
			}

			query = query.OrderBy(fmt.Sprintf("p.%s %s", o.Column, o.Order))
		}
	} else {
		query = query.OrderBy("p.created_at DESC")
	}

	sql, args, err := query.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var posts []PostWithAuthor
	err = db.SelectContext(ctx, &posts, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}

	return posts, nil
}
