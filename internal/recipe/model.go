package recipe

import (
	"time"
)

type Recipe struct {
	ID          string         `db:"recipe_id"   json:"recipe_id"`
	User        string         `db:"user_id"     json:"user_id"`
	Title       string         `db:"title"       json:"title"`
	Description *string        `db:"description" json:"description"`
	PhotoURL    *string        `db:"photo_url"   json:"photo_url"`
	SourceType  *SourceType    `db:"source_type" json:"source_type"`
	Source      *string        `db:"source"      json:"source"`
	Components  []Component    `db:"-"           json:"components"`
	PrepTime    *time.Duration `db:"prep_time"   json:"prep_time"`
	CookTime    *time.Duration `db:"cook_time"   json:"cook_time"`
	Serves      *uint32        `db:"serves"      json:"serves"`
	Cuisine     *Cuisine       `db:"cuisine"     json:"cuisine"`
	Category    *Category      `db:"category"    json:"category"`
	CreatedAt   time.Time      `db:"created_at"  json:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"  json:"updated_at"`
	DeletedAt   *time.Time     `db:"deleted_at"  json:"deleted_at,omitempty"`
}

type Component struct {
	ID           string       `db:"component_id" json:"component_id"`
	Recipe       string       `db:"recipe_id"    json:"recipe_id,omitempty"`
	Name         *string      `db:"name"         json:"name"`
	Position     int          `db:"position"     json:"position"`
	Ingredients  []Ingredient `db:"-"            json:"ingredients"`
	Instructions []string     `db:"-"            json:"instructions"`
}

type SourceType string

const (
	SourceURL      SourceType = "url"
	SourceBook     SourceType = "book"
	SourcePersonal SourceType = "personal"
)

func (s SourceType) Valid() bool {
	switch s {
	case SourceURL, SourceBook, SourcePersonal:
		return true
	default:
		return false
	}
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
	// count
	Whole Unit = "whole"
)

func (u Unit) Valid() bool {
	switch u {
	case Tsp, Tbsp, Cup, FlOz, Pt, Qt, Gal, Oz, Lb, G, Kg, Ml, L, Pinch, Dash, Whole:
		return true
	default:
		return false
	}
}

type Ingredient struct {
	ID        string  `db:"ingredient_id" json:"ingredient_id,omitempty"`
	Component string  `db:"component_id"  json:"component_id,omitempty"`
	Amount    float64 `db:"amount"        json:"amount"`
	Unit      Unit    `db:"unit"          json:"unit"`
	Item      string  `db:"item"          json:"item"`
}

type Instruction struct {
	ID          string `db:"instruction_id" json:"instruction_id,omitempty"`
	Component   string `db:"component_id"   json:"component_id,omitempty"`
	StepNumber  int32  `db:"step_number"    json:"step_number"`
	Instruction string `db:"instruction"    json:"instruction"`
}

type EditableFields struct {
	Title       *string             `json:"title"`
	Description *string             `json:"description"`
	PhotoURL    *string             `json:"photo_url"`
	SourceType  *SourceType         `json:"source_type"`
	Source      *string             `json:"source"`
	Components  *[]ComponentRequest `json:"components"`
	PrepTime    *time.Duration      `json:"prep_time"`
	CookTime    *time.Duration      `json:"cook_time"`
	Serves      *uint32             `json:"serves"`
	Cuisine     *Cuisine            `json:"cuisine"`
	Category    *Category           `json:"category"`
}
