package recipe

import (
	"time"
)

type Recipe struct {
	Id           string         `db:"recipe_id"   json:"recipe_id"`
	User         string         `db:"user_id"     json:"user_id"`
	Title        string         `db:"title"       json:"title"`
	Description  *string        `db:"description" json:"description"`
	PhotoUrl     *string        `db:"photo_url"   json:"photo_url"`
	SourceUrl    *string        `db:"source_url"  json:"source_url"`
	Ingredients  []Ingredient   `db:"-"           json:"ingredients"`
	Instructions []string       `db:"-"           json:"instructions"`
	CookTime     *time.Duration `db:"cook_time"   json:"cook_time"`
	Serves       *uint32        `db:"serves"      json:"serves"`
	Cuisine      *Cuisine       `db:"cuisine"     json:"cuisine"`
	Category     *Category      `db:"category"    json:"category"`
	CreatedAt    time.Time      `db:"created_at"  json:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"  json:"updated_at"`
	DeletedAt    *time.Time     `db:"deleted_at"  json:"deleted_at,omitempty"`
}

type RecipeRow struct {
	Recipe
	RawIngredients  []byte `db:"ingredients_json"`
	RawInstructions []byte `db:"instructions_json"`
}

type Cuisine string

const (
	American   Cuisine = "American"
	Chinese    Cuisine = "Chinese"
	French     Cuisine = "French"
	Indian     Cuisine = "Indian"
	Italian    Cuisine = "Italian"
	Japanese   Cuisine = "Japanese"
	Vietnamese Cuisine = "Vietnamese"
)

func (c Cuisine) Valid() bool {
	switch c {
	case American, Chinese, French, Indian, Italian, Japanese, Vietnamese:
		return true
	default:
		return false
	}
}

type Category string

const (
	Appetizer Category = "Appetizer"
	Main      Category = "Main"
	Dessert   Category = "Dessert"
	Beverage  Category = "Beverage"
	Side      Category = "Side"
)

func (c Category) Valid() bool {
	switch c {
	case Appetizer, Main, Dessert, Beverage, Side:
		return true
	default:
		return false
	}
}

type Unit string

const (
	// volume
	Tsp   Unit = "tsp"
	Tbsp  Unit = "tbsp"
	Cup   Unit = "cup"
	FlOz  Unit = "fl oz"
	Pt    Unit = "pt"
	Qt    Unit = "qt"
	Gal   Unit = "gal"
	Ml    Unit = "ml"
	L     Unit = "l"
	Pinch Unit = "pinch"
	Dash  Unit = "dash"
	// weight
	Oz Unit = "oz"
	Lb Unit = "lb"
	G  Unit = "g"
	Kg Unit = "kg"
)

func (u Unit) Valid() bool {
	switch u {
	case Tsp, Tbsp, Cup, FlOz, Pt, Qt, Gal, Oz, Lb, G, Kg, Ml, L, Pinch, Dash:
		return true
	default:
		return false
	}
}

type Ingredient struct {
	Id       string  `db:"ingredient_id" json:"ingredient_id,omitempty"`
	RecipeId string  `db:"recipe_id"     json:"recipe_id,omitempty"`
	Amount   float64 `db:"amount"        json:"amount"`
	Unit     Unit    `db:"unit"          json:"unit"`
	Item     string  `db:"item"          json:"item"`
}

type Instruction struct {
	Id          string `db:"instruction_id" json:"instruction_id,omitempty"`
	RecipeId    string `db:"recipe_id"      json:"recipe_id,omitempty"`
	StepNumber  int32  `db:"step_number"    json:"step_number"`
	Instruction string `db:"instruction"    json:"instruction"`
}

type EditableFields struct {
	Title        *string        `json:"title"`
	Description  *string        `json:"description"`
	PhotoUrl     *string        `json:"photo_url"`
	SourceUrl    *string        `json:"source_url"`
	Ingredients  *[]Ingredient  `json:"ingredients"`
	Instructions *[]string      `json:"instructions"`
	CookTime     *time.Duration `json:"cook_time"`
	Serves       *uint32        `json:"serves"`
	Cuisine      *Cuisine       `json:"cuisine"`
	Category     *Category      `json:"category"`
}
