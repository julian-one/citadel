package post

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"citadel/internal/database"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type ListOptions struct {
	Search         string
	User           string
	Public         *bool
	IncludeDeleted bool
	OrderBy        []database.Order
	Pagination     database.Pagination
}

func ParseListOptions(r *http.Request, user string) (ListOptions, error) {
	q := r.URL.Query()

	var opts ListOptions

	if search := q.Get("search"); search != "" {
		opts.Search = search
	}
	opts.User = user

	if public := q.Get("public"); public != "" {
		b, err := strconv.ParseBool(public)
		if err != nil {
			return opts, fmt.Errorf("invalid public value: %s (must be 'true' or 'false')", public)
		}
		opts.Public = &b
	}

	parsed, err := database.ParseOrder(q.Get("order_by"))
	if err != nil {
		return opts, err
	}
	opts.OrderBy = parsed

	pagination, err := database.ParsePagination(q.Get("limit"), q.Get("offset"))
	if err != nil {
		return opts, err
	}
	opts.Pagination = pagination

	return opts, opts.validate()
}

func (opts ListOptions) validate() error {
	if len(opts.OrderBy) == 0 {
		return nil
	}

	t := reflect.TypeFor[Post]()
	validColumns := make(map[string]bool, t.NumField())
	for field := range t.Fields() {
		if tag := field.Tag.Get("db"); tag != "" && tag != "-" {
			validColumns[tag] = true
		}
	}

	for _, o := range opts.OrderBy {
		if !validColumns[o.Column] {
			return fmt.Errorf("invalid column name: %s", o.Column)
		}
	}
	return nil
}

func applyFilters(q sq.SelectBuilder, opts ListOptions) sq.SelectBuilder {
	if !opts.IncludeDeleted {
		q = q.Where("p.deleted_at IS NULL")
	}

	if opts.Search != "" {
		searchPattern := "%" + strings.ToLower(opts.Search) + "%"
		q = q.Where(sq.Expr("LOWER(p.title) LIKE ?", searchPattern))
	}

	if opts.User != "" && opts.Public == nil {
		q = q.Where(sq.Or{
			sq.Eq{"p.user_id": opts.User},
			sq.Eq{"p.public": true},
		})
	} else {
		if opts.User != "" {
			q = q.Where(sq.Eq{"p.user_id": opts.User})
		}
		if opts.Public != nil {
			q = q.Where(sq.Eq{"p.public": *opts.Public})
		}
	}

	return q
}

func Count(ctx context.Context, db *sqlx.DB, opts ListOptions) (int, error) {
	q := applyFilters(database.QB.Select("COUNT(*)").From("posts p"), opts)

	query, args, err := q.ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build count query: %w", err)
	}

	var total int
	if err = db.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("failed to count posts: %w", err)
	}

	return total, nil
}

func List(ctx context.Context, db *sqlx.DB, opts ListOptions) ([]PostWithAuthor, error) {
	q := applyFilters(
		database.QB.Select("p.*, u.email, u.username").
			From("posts p").
			InnerJoin("users u ON (u.user_id = p.user_id)"),
		opts,
	)

	if len(opts.OrderBy) > 0 {
		for _, o := range opts.OrderBy {
			q = q.OrderBy(fmt.Sprintf("p.%s %s", o.Column, o.Direction))
		}
	} else {
		q = q.OrderBy("p.created_at DESC")
	}

	q = q.Limit(uint64(opts.Pagination.Limit)).Offset(uint64(opts.Pagination.Offset))

	query, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	posts := []PostWithAuthor{}
	if err = db.SelectContext(ctx, &posts, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}
	if posts == nil {
		posts = []PostWithAuthor{}
	}

	return posts, nil
}
