package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/zeebo/errs"
	"kitchen_nerd/pkg/util"
	"time"

	"kitchen_nerd/recipes"
)

var _ recipes.DB = (*recipesDB)(nil)

var ErrRecipes = errs.Class("recipes repository error")

// recipesDB provides access to users db.
//
// architecture: Database.
type recipesDB struct {
	pool *pgxpool.Pool
}

func (s *recipesDB) CreateIngredient(ctx context.Context, ingredient recipes.RecipeIngredient) error {
	query := `INSERT INTO recipe_ingredients (id, name, recipe_id, quantity, unit, optional) 
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := s.pool.Exec(ctx, query, uuid.New(), ingredient.Name, ingredient.RecipeID, ingredient.Quantity, ingredient.Unit, ingredient.Optional)
	return err
}

func (s *recipesDB) Count(ctx context.Context) (uint64, error) {
	query := `SELECT COUNT(id) FROM recipes`
	var count uint64
	if err := s.pool.QueryRow(ctx, query).Scan(&count); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, recipes.ErrNoRecipe.New("")
		}
		return 0, err
	}

	return count, nil
}

// List returns all recipes from the database.
func (s *recipesDB) List(ctx context.Context, pagination *util.PaginationReq) ([]*recipes.Recipe, error) {
	query := `SELECT id, title, photo, description, instructions, created_at 
	          FROM recipes 
	          LIMIT $1 
	          OFFSET $2`
	rows, err := s.pool.Query(ctx, query, pagination.Size, pagination.GetDBOffset())
	if err != nil {
		return nil, ErrRecipes.Wrap(err)
	}
	defer rows.Close()

	var list []*recipes.Recipe
	for rows.Next() {

		recipe := new(recipes.Recipe)
		if err = rows.Scan(&recipe.ID, &recipe.Title, &recipe.PhotoBase64, &recipe.Description, &recipe.Instructions, &recipe.CreatedAt); err != nil {
			return nil, ErrRecipes.Wrap(err)
		}

		ingredients, err := s.GetIngredients(ctx, recipe.ID)
		if err != nil {
			return nil, ErrRecipes.Wrap(err)
		}

		recipe.Ingredients = ingredients
		list = append(list, recipe)
	}

	return list, nil
}

// GetIngredients returns a list of ingredients by recipe ID.
func (s *recipesDB) GetIngredients(ctx context.Context, recipeID uuid.UUID) ([]recipes.RecipeIngredient, error) {
	query := `SELECT id, name, recipe_id, quantity, unit, optional
	          FROM recipe_ingredients
	          WHERE recipe_id = $1`
	rows, err := s.pool.Query(ctx, query, recipeID)
	if err != nil {
		return nil, ErrRecipes.Wrap(err)
	}

	defer rows.Close()
	var ingredients []recipes.RecipeIngredient

	for rows.Next() {
		var ingredient recipes.RecipeIngredient
		if err = rows.Scan(&ingredient.ID, &ingredient.Name, &ingredient.RecipeID, &ingredient.Quantity, &ingredient.Unit, &ingredient.Optional); err != nil {
			return nil, ErrRecipes.Wrap(err)
		}

		ingredients = append(ingredients, ingredient)
	}

	return ingredients, nil
}

// GetRecipe returns a recipe by its ID from the database.
func (s *recipesDB) GetRecipe(ctx context.Context, id uuid.UUID) (*recipes.Recipe, error) {
	var recipe *recipes.Recipe
	err := s.pool.QueryRow(ctx, "SELECT * FROM recipes WHERE id = $1", id).Scan(&recipe.ID, &recipe.Title, &recipe.PhotoBase64, &recipe.Description, &recipe.Instructions, &recipe.CreatedAt)
	if err != nil {
		return nil, ErrRecipes.Wrap(err)
	}

	return recipe, nil
}

// CreateRecipe adds a new recipe to the database.
func (s *recipesDB) CreateRecipe(ctx context.Context, recipe *recipes.Recipe) error {
	recipe.CreatedAt = time.Now()

	_, err := s.pool.Exec(ctx, "INSERT INTO recipes (id, title, photo, description, instructions, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		recipe.ID, recipe.Title, recipe.PhotoBase64, recipe.Description, recipe.Instructions, recipe.CreatedAt)
	if err != nil {
		return ErrRecipes.Wrap(err)
	}

	return nil
}

// UpdateRecipe updates a recipe in the database.
func (s *recipesDB) UpdateRecipe(ctx context.Context, updatedRecipe *recipes.Recipe) error {
	var existingRecipe *recipes.Recipe
	err := s.pool.QueryRow(ctx, "SELECT * FROM recipes WHERE id = $1", updatedRecipe.ID).Scan(&existingRecipe.ID, &existingRecipe.Title, &existingRecipe.PhotoBase64, &existingRecipe.Description, &existingRecipe.Instructions, &existingRecipe.CreatedAt)
	if err != nil {
		return ErrRecipes.Wrap(err)
	}

	// Update only non-empty fields
	if updatedRecipe.Title != "" {
		existingRecipe.Title = updatedRecipe.Title
	}
	if updatedRecipe.PhotoBase64 != "" {
		existingRecipe.PhotoBase64 = updatedRecipe.PhotoBase64
	}
	if updatedRecipe.Description != "" {
		existingRecipe.Description = updatedRecipe.Description
	}
	if updatedRecipe.Instructions != "" {
		existingRecipe.Instructions = updatedRecipe.Instructions
	}

	_, err = s.pool.Exec(context.TODO(), "UPDATE recipes SET title = $2, photo = $3, description = $4, instructions = $5 WHERE id = $1",
		existingRecipe.ID, existingRecipe.Title, existingRecipe.PhotoBase64, existingRecipe.Description, existingRecipe.Instructions)
	if err != nil {
		return ErrRecipes.Wrap(err)
	}

	return nil
}

// DeleteRecipe deletes a recipe from the database.
func (s *recipesDB) DeleteRecipe(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM recipes WHERE id = $1", id)
	if err != nil {
		return ErrRecipes.Wrap(err)
	}

	return nil
}
