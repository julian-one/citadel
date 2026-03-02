package recipe

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"citadel/internal/database"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

func isValidColumn(column string) bool {
	validColumns := map[string]bool{
		"recipe_id":  true,
		"user_id":    true,
		"title":      true,
		"cuisine":    true,
		"created_at": true,
		"updated_at": true,
	}
	return validColumns[column]
}

type ListOptions struct {
	Search  string
	User    string
	OrderBy []database.OrderBy
}

func ParseListOptions(r *http.Request, userId string) (ListOptions, error) {
	query := r.URL.Query()

	var opts ListOptions

	if search := query.Get("search"); search != "" {
		opts.Search = search
	}
	opts.User = userId

	if orderBy := query.Get("order_by"); orderBy != "" {
		parsed, err := database.ParseOrderBy(orderBy)
		if err != nil {
			return opts, err
		}
		opts.OrderBy = parsed
	}

	return opts, nil
}

func List(ctx context.Context, db *sqlx.DB, opts ListOptions) ([]Recipe, error) {
	queryBuilder := sq.Select("r.*").
		Column("(SELECT json_group_array(json_object('amount', amount, 'unit', unit, 'item', item)) FROM ingredients WHERE recipe_id = r.recipe_id) AS ingredients_json").
		Column("(SELECT json_group_array(instruction) FROM instructions WHERE recipe_id = r.recipe_id ORDER BY step_number ASC) AS instructions_json").
		From("recipes r").
		Where("r.deleted_at IS NULL").
		Where(sq.Eq{"r.user_id": opts.User})

	if opts.Search != "" {
		searchPattern := "%" + strings.ToLower(opts.Search) + "%"
		queryBuilder = queryBuilder.Where(sq.Expr("LOWER(r.title) LIKE ?", searchPattern))
	}

	if len(opts.OrderBy) > 0 {
		for _, o := range opts.OrderBy {
			if !o.Order.Valid() {
				return nil, fmt.Errorf("invalid order direction: %s", o.Order)
			}

			if !isValidColumn(o.Column) {
				return nil, fmt.Errorf("invalid column name: %s", o.Column)
			}

			queryBuilder = queryBuilder.OrderBy(fmt.Sprintf("r.%s %s", o.Column, o.Order))
		}
	} else {
		queryBuilder = queryBuilder.OrderBy("r.created_at DESC")
	}

	sql, args, err := queryBuilder.PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var rows []RecipeRow
	err = db.SelectContext(ctx, &rows, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list recipes: %w", err)
	}

	recipes := make([]Recipe, len(rows))
	for i, row := range rows {
		if len(row.RawIngredients) > 0 {
			if err := json.Unmarshal(row.RawIngredients, &row.Recipe.Ingredients); err != nil {
				return nil, fmt.Errorf(
					"failed to unmarshal ingredients for recipe %s: %w",
					row.Recipe.Id,
					err,
				)
			}
		}

		if len(row.RawInstructions) > 0 {
			if err := json.Unmarshal(row.RawInstructions, &row.Recipe.Instructions); err != nil {
				return nil, fmt.Errorf(
					"failed to unmarshal instructions for recipe %s: %w",
					row.Recipe.Id,
					err,
				)
			}
		}
		recipes[i] = row.Recipe
	}

	return recipes, nil
}
