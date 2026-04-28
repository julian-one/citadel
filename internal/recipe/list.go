package recipe

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
	Search     string
	Cuisine    string
	Category   string
	Bookmarks  bool
	User       string
	Deleted    bool
	OrderBy    []database.Order
	Pagination database.Pagination
}

func ParseListOptions(r *http.Request, user string) (ListOptions, error) {
	query := r.URL.Query()

	var opts ListOptions
	opts.User = user

	if search := query.Get("search"); search != "" {
		opts.Search = search
	}

	if cuisine := query.Get("cuisine"); cuisine != "" {
		opts.Cuisine = cuisine
	}

	if category := query.Get("category"); category != "" {
		opts.Category = category
	}

	if bookmarks := query.Get("bookmarks"); bookmarks != "" {
		b, err := strconv.ParseBool(bookmarks)
		if err != nil {
			return opts, fmt.Errorf(
				"invalid bookmarks value: %s (must be 'true' or 'false')",
				bookmarks,
			)
		}
		opts.Bookmarks = b
	}

	parsed, err := database.ParseOrder(query.Get("order_by"))
	if err != nil {
		return opts, err
	}
	opts.OrderBy = parsed

	pagination, err := database.ParsePagination(query.Get("limit"), query.Get("offset"))
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

	t := reflect.TypeFor[Recipe]()
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
	if !opts.Deleted {
		q = q.Where("r.deleted_at IS NULL")
	}

	if opts.Search != "" {
		searchPattern := "%" + strings.ToLower(opts.Search) + "%"
		q = q.Where(sq.Expr("LOWER(r.title) LIKE ?", searchPattern))
	}
	if opts.Cuisine != "" {
		q = q.Where(sq.Eq{"r.cuisine": opts.Cuisine})
	}
	if opts.Category != "" {
		q = q.Where(sq.Eq{"r.category": opts.Category})
	}
	if opts.Bookmarks && opts.User != "" {
		q = q.InnerJoin(
			"recipe_bookmarks b ON (b.recipe_id = r.recipe_id AND b.user_id = ?)",
			opts.User,
		)
	}

	return q
}

func Count(ctx context.Context, db sqlx.QueryerContext, opts ListOptions) (int, error) {
	q := applyFilters(database.QB.Select("COUNT(*)").From("recipes r"), opts)

	query, args, err := q.ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build count query: %w", err)
	}

	var total int
	if err = db.QueryRowxContext(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("failed to count recipes: %w", err)
	}

	return total, nil
}

func List(ctx context.Context, db sqlx.QueryerContext, opts ListOptions) ([]Recipe, error) {
	q := applyFilters(
		database.QB.Select("r.*").
			From("recipes r"),
		opts,
	)

	if len(opts.OrderBy) > 0 {
		for _, o := range opts.OrderBy {
			q = q.OrderBy(fmt.Sprintf("r.%s %s", o.Column, o.Direction))
		}
	} else {
		q = q.OrderBy("r.created_at DESC")
	}

	q = q.Limit(uint64(opts.Pagination.Limit)).Offset(uint64(opts.Pagination.Offset))

	query, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	recipes := make([]Recipe, 0)
	if err = sqlx.SelectContext(ctx, db, &recipes, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list recipes: %w", err)
	}

	recipe_ids := make([]string, len(recipes))
	for i, r := range recipes {
		recipe_ids[i] = r.ID
	}

	all_components, err := LoadComponents(ctx, db, recipe_ids)
	if err != nil {
		return nil, err
	}

	by_recipe := make(map[string][]Component, len(all_components))
	for _, c := range all_components {
		by_recipe[c.Recipe] = append(by_recipe[c.Recipe], c)
	}

	for i := range recipes {
		recipes[i].Components = by_recipe[recipes[i].ID]
		if recipes[i].Components == nil {
			recipes[i].Components = []Component{}
		}
	}

	return recipes, nil
}
