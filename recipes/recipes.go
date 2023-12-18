package recipes

import (
	"context"
	"github.com/google/uuid"
	"github.com/zeebo/errs"
	"kitchen_nerd/pkg/util"
	"time"
)

var ErrNoRecipe = errs.Class("recipe does not exist")

type DB interface {
	CreateRecipe(ctx context.Context, recipe *Recipe) error
	List(ctx context.Context, pagination *util.PaginationReq) ([]*Recipe, error)
	Count(ctx context.Context) (uint64, error)
	GetRecipe(ctx context.Context, id uuid.UUID) (*Recipe, error)
	UpdateRecipe(ctx context.Context, updatedRecipe *Recipe) error
	DeleteRecipe(ctx context.Context, id uuid.UUID) error
	CreateIngredient(ctx context.Context, ingredient RecipeIngredient) error
}

type Recipe struct {
	ID           uuid.UUID `json:"id"`
	Title        string    `json:"title"`
	PhotoBase64  string    `json:"photoBase64"`
	Description  string    `json:"description"`
	Instructions string    `json:"instructions"`
	CreatedAt    time.Time `json:"createdAt"`

	Ingredients []RecipeIngredient `json:"ingredients"`
}

func NewRecipe(title, photo, description, instructions string) *Recipe {
	return &Recipe{
		ID:           uuid.New(),
		Title:        title,
		PhotoBase64:  photo,
		Description:  description,
		Instructions: instructions,
		CreatedAt:    time.Now(),
	}
}

// UnitType represents the unit of measurement.
type UnitType string

// Possible values for unit of measurement.
const (
	Gram       UnitType = "gram"
	Milliliter UnitType = "milliliter"
	Teaspoon   UnitType = "teaspoon"
	Tablespoon UnitType = "tablespoon"
	Piece      UnitType = "piece" // Non-standard unit of measurement
)

// IsValidUnit checks if the value of string belongs to UnitType.
func IsValidUnit(unit string) bool {
	switch UnitType(unit) {
	case Gram, Milliliter, Teaspoon, Tablespoon, Piece:
		return true
	}
	return false
}

type IngredientForRecipe struct {
	RecipeID     uuid.UUID `json:"recipeID"`
	IngredientID uuid.UUID `json:"ingredientID"`
	Name         string    `json:"name"`
	Quantity     int       `json:"quantity"`
	Unit         string    `json:"unit"`
	Optional     bool      `json:"optional"`
}

type RecipeIngredient struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	RecipeID uuid.UUID `json:"recipeID"`
	Quantity float64   `json:"quantity"`
	Optional bool      `json:"optional"`
	Unit     string    `json:"unit"`
}
