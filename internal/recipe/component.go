package recipe

import (
	"context"
	"fmt"

	"citadel/internal/database"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

// LoadComponents queries the recipe_components table for the given recipe IDs
// and eagerly loads their nested ingredients and instructions.
func LoadComponents(
	ctx context.Context,
	db sqlx.QueryerContext,
	recipeIds []string,
) ([]Component, error) {
	if len(recipeIds) == 0 {
		return []Component{}, nil
	}

	ids := make([]any, len(recipeIds))
	for i, id := range recipeIds {
		ids[i] = id
	}

	// Load components.
	var components []Component

	query, args, err := database.QB.
		Select("*").
		From("recipe_components").
		Where(sq.Eq{"recipe_id": ids}).
		OrderBy("position").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build components query: %w", err)
	}

	if err = sqlx.SelectContext(ctx, db, &components, query, args...); err != nil {
		return nil, fmt.Errorf("failed to load components: %w", err)
	}

	if len(components) == 0 {
		return components, nil
	}

	compIds := make([]any, len(components))
	for i, c := range components {
		compIds[i] = c.ID
	}

	// Load ingredients.
	var ingredients []Ingredient

	query, args, err = database.QB.
		Select("*").
		From("ingredients").
		Where(sq.Eq{"component_id": compIds}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build ingredients query: %w", err)
	}

	if err = sqlx.SelectContext(ctx, db, &ingredients, query, args...); err != nil {
		return nil, fmt.Errorf("failed to load ingredients: %w", err)
	}

	ingredientsByComp := make(map[string][]Ingredient, len(components))
	for _, ing := range ingredients {
		ingredientsByComp[ing.Component] = append(ingredientsByComp[ing.Component], ing)
	}

	for i := range components {
		components[i].Ingredients = ingredientsByComp[components[i].ID]
		if components[i].Ingredients == nil {
			components[i].Ingredients = []Ingredient{}
		}
	}

	// Load instructions.
	var instructions []Instruction

	query, args, err = database.QB.
		Select("*").
		From("instructions").
		Where(sq.Eq{"component_id": compIds}).
		OrderBy("step_number").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build instructions query: %w", err)
	}

	if err = sqlx.SelectContext(ctx, db, &instructions, query, args...); err != nil {
		return nil, fmt.Errorf("failed to load instructions: %w", err)
	}

	instructionsByComp := make(map[string][]string, len(components))
	for _, ins := range instructions {
		instructionsByComp[ins.Component] = append(
			instructionsByComp[ins.Component],
			ins.Instruction,
		)
	}

	for i := range components {
		components[i].Instructions = instructionsByComp[components[i].ID]
		if components[i].Instructions == nil {
			components[i].Instructions = []string{}
		}
	}

	return components, nil
}
