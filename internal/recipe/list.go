package recipe

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"citadel/internal/database"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type ListOptions struct {
	Search         string
	Cuisine        string
	Category       string
	IncludeDeleted bool
	OrderBy        []database.Order
	Pagination     database.Pagination
}

func ParseListOptions(r *http.Request) (ListOptions, error) {
	query := r.URL.Query()

	var opts ListOptions

	if search := query.Get("search"); search != "" {
		opts.Search = search
	}

	if cuisine := query.Get("cuisine"); cuisine != "" {
		opts.Cuisine = cuisine
	}

	if category := query.Get("category"); category != "" {
		opts.Category = category
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
	if !opts.IncludeDeleted {
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

	return q
}

func Count(ctx context.Context, db *sqlx.DB, opts ListOptions) (int, error) {
	q := applyFilters(database.QB.Select("COUNT(*)").From("recipes r"), opts)

	query, args, err := q.ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build count query: %w", err)
	}

	var total int
	if err = db.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("failed to count recipes: %w", err)
	}

	return total, nil
}

func List(ctx context.Context, db *sqlx.DB, opts ListOptions) ([]Recipe, error) {
	q := applyFilters(
		database.QB.Select("r.*").
			Column("(SELECT json_group_array(json_object('amount', amount, 'unit', unit, 'item', item)) FROM ingredients WHERE recipe_id = r.recipe_id) AS ingredients_json").
			Column("(SELECT json_group_array(instruction) FROM instructions WHERE recipe_id = r.recipe_id ORDER BY step_number ASC) AS instructions_json").
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

	var rows []recipeRow
	if err = db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list recipes: %w", err)
	}

	recipes := make([]Recipe, len(rows))
	for i, row := range rows {
		if len(row.RawIngredients) > 0 {
			if err := json.Unmarshal(row.RawIngredients, &row.Ingredients); err != nil {
				return nil, fmt.Errorf(
					"failed to unmarshal ingredients for recipe %s: %w",
					row.ID,
					err,
				)
			}
		}
		if len(row.RawInstructions) > 0 {
			if err := json.Unmarshal(row.RawInstructions, &row.Instructions); err != nil {
				return nil, fmt.Errorf(
					"failed to unmarshal instructions for recipe %s: %w",
					row.ID,
					err,
				)
			}
		}
		recipes[i] = row.Recipe
	}

	return recipes, nil
}
