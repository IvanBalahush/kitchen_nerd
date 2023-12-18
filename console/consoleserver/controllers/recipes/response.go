package recipes_controller

import (
	"kitchen_nerd/pkg/util"
	"kitchen_nerd/recipes"
)

type ListResponse struct {
	Recipes            []*recipes.Recipe        `json:"recipes"`
	PaginationResponse *util.PaginationResponse `json:"pagination"`
}

func NewListResponse(recipes []*recipes.Recipe, paginationResponse *util.PaginationResponse) *ListResponse {
	return &ListResponse{Recipes: recipes, PaginationResponse: paginationResponse}
}
