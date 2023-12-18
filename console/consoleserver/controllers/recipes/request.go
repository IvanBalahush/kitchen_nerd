package recipes_controller

import "kitchen_nerd/recipes"

type RecipeRequest struct {
	Title        string                     `json:"title"`
	PhotoBase64  string                     `json:"photoBase64"`
	Description  string                     `json:"description"`
	Instructions string                     `json:"instructions"`
	Ingredients  []recipes.RecipeIngredient `json:"ingredients"`
}

//type Recipe struct {
//	Title        string `json:"title"`
//	PhotoBase64  string `json:"photoBase64"`
//	Description  string `json:"description"`
//	Instructions string `json:"instructions"`
//	Ingredients  []struct {
//		Name     string `json:"name"`
//		Quantity string `json:"quantity"`
//		Unit     string `json:"unit"`
//		Optional bool   `json:"optional"`
//	} `json:"ingredients"`
//}
