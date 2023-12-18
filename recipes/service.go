package recipes

import (
	"context"
	"github.com/google/uuid"
	"github.com/zeebo/errs"
	"kitchen_nerd/pkg/util"
)

var ErrRecipes = errs.Class("recipes service error")

// Service is handling recipes related logic.
//
// architecture: Service
type Service struct {
	recipes DB
}

func NewService(recipes DB) *Service {
	return &Service{recipes: recipes}
}

func (service *Service) CreateIngredients(ctx context.Context, recipeID uuid.UUID, ingredients []RecipeIngredient) error {
	for _, ingredient := range ingredients {
		ingredient.RecipeID = recipeID
		err := service.recipes.CreateIngredient(ctx, ingredient)
		if err != nil {
			return err
		}
	}
	return nil
}

func (service *Service) Create(ctx context.Context, recipe *Recipe, ingredients []RecipeIngredient) error {
	recipe.Ingredients = ingredients
	err := service.recipes.CreateRecipe(ctx, recipe)
	if err != nil {
		return err
	}

	err = service.CreateIngredients(ctx, recipe.ID, recipe.Ingredients)

	return err
}

func (service *Service) List(ctx context.Context, pagination *util.PaginationReq) ([]*Recipe, uint64, error) {
	count, err := service.recipes.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	if count == 0 {
		return make([]*Recipe, 0), 0, nil
	}

	list, err := service.recipes.List(ctx, pagination)
	if err != nil {
		return nil, 0, err
	}

	return list, count, err
}

func (service *Service) Get(ctx context.Context, id uuid.UUID) (*Recipe, error) {
	recipe, err := service.recipes.GetRecipe(ctx, id)
	return recipe, err
}

func (service *Service) Delete(ctx context.Context, id uuid.UUID) error {
	err := service.recipes.DeleteRecipe(ctx, id)
	return err
}

func (service *Service) Update(ctx context.Context, id uuid.UUID, title, photo, description, instructions string) error {

	updatedRecipe := &Recipe{
		ID:           id,
		Title:        title,
		PhotoBase64:  photo,
		Description:  description,
		Instructions: instructions,
	}

	err := service.recipes.UpdateRecipe(ctx, updatedRecipe)
	if err != nil {
		return err
	}

	return nil
}
