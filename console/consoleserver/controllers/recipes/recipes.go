package recipes_controller

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/zeebo/errs"
	"html/template"
	"kitchen_nerd/pkg/util"
	"kitchen_nerd/recipes"
	"log"
	"net/http"
)

var (
	// ErrRecipes is an internal error type for recipes controller.
	ErrRecipes = errs.Class("recipes controller error")
)

type Templates struct {
	List   *template.Template
	Get    *template.Template
	Create *template.Template
}

// Recipes is a mvc controller that handles all recipes related views.
type Recipes struct {
	recipes *recipes.Service

	templates Templates
}

func NewRecipes(recipes *recipes.Service, templates Templates) *Recipes {
	recipesController := &Recipes{
		recipes:   recipes,
		templates: templates,
	}

	return recipesController
}

func (c *Recipes) List(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if err := c.templates.List.Execute(w, nil); err != nil {
			http.Error(w, "could not execute recipes list template", http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		ctx := r.Context()
		pagination := util.NewPaginationReq(10, 1)
		if err := pagination.ProcessQueryParams(r.URL.Query()); err != nil {
			c.serveError(w, http.StatusInternalServerError, err)
			return
		}

		res, total, err := c.recipes.List(ctx, pagination)
		if err != nil {
			c.serveError(w, http.StatusInternalServerError, err)
			return
		}
		response := NewListResponse(res, util.NewPaginationResponse(pagination.Size, pagination.Page, total))
		if err = json.NewEncoder(w).Encode(response); err != nil {
			log.Println("failed to write json error response", ErrRecipes.Wrap(err))
			return
		}
	}
}

func (c *Recipes) Create(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if err := c.templates.Create.Execute(w, nil); err != nil {
			http.Error(w, "could not execute recipes list template", http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		ctx := r.Context()
		var request *RecipeRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Println(err)
			c.serveError(w, http.StatusBadRequest, ErrRecipes.Wrap(err))
			return
		}

		recipe := recipes.NewRecipe(request.Title, request.PhotoBase64, request.Description, request.Instructions)

		err := c.recipes.Create(ctx, recipe, request.Ingredients)
		if err != nil {
			c.serveError(w, http.StatusInternalServerError, ErrRecipes.Wrap(err))
			return
		}

		//decoder := json.NewDecoder(r.Body)
		//var recipe Recipe
		//
		//err := decoder.Decode(&recipe)
		//if err != nil {
		//	http.Error(w, "Invalid JSON", http.StatusBadRequest)
		//	log.Println(err)
		//	return
		//}
		//
		//// Теперь у вас есть структура recipe, которую вы можете использовать по своему усмотрению.
		//// Например, вы можете сохранить ее в базу данных или выполнить другие операции.
		//
		//// Выведем полученные данные для примера.
		//fmt.Printf("Recipe Title: %s\n", recipe.Title)
		//fmt.Printf("Photo Base64: %s\n", recipe.PhotoBase64)
		//fmt.Printf("Description: %s\n", recipe.Description)
		//fmt.Printf("Instructions: %s\n", recipe.Instructions)
		//
		//for i, ingredient := range recipe.Ingredients {
		//	fmt.Printf("Ingredient %d:\n", i+1)
		//	fmt.Printf("  Name: %s\n", ingredient.Name)
		//	fmt.Printf("  Quantity: %d\n", ingredient.Quantity)
		//	fmt.Printf("  Unit: %s\n", ingredient.Unit)
		//	fmt.Printf("  Optional: %v\n", ingredient.Optional)
		//}
		//
		//// Возвращаем успешный статус
		//w.WriteHeader(http.StatusOK)
	}
}

func (c *Recipes) Get(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if err := c.templates.Get.Execute(w, nil); err != nil {
			http.Error(w, "could not execute recipes list template", http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		ctx := r.Context()
		vars := mux.Vars(r)
		id, err := uuid.Parse(vars["id"])
		if err != nil {
			c.serveError(w, http.StatusBadRequest, err)
			return
		}

		recipe, err := c.recipes.Get(ctx, id)
		if err != nil {
			c.serveError(w, http.StatusBadRequest, err)
			return
		}
		if err = json.NewEncoder(w).Encode(recipe); err != nil {
			log.Println("failed to write json error response", ErrRecipes.Wrap(err))
			return
		}
	}
}

//
//func (c *Recipes) Update(w http.ResponseWriter, r *http.Request) {
//	switch r.Method {
//	case http.MethodGet:
//		if err := c.templates.Get.Execute(w, nil); err != nil {
//			http.Error(w, "could not execute recipes list template", http.StatusInternalServerError)
//			return
//		}
//	case http.MethodPost:
//		ctx := r.Context()
//		vars := mux.Vars(r)
//		id, err := uuid.Parse(vars["id"])
//		if err != nil {
//			c.serveError(w, http.StatusBadRequest, err)
//			return
//		}
//		var request CreateRequest
//		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
//			c.serveError(w, http.StatusBadRequest, ErrRecipes.Wrap(err))
//			return
//		}
//
//		err = c.recipes.Update(ctx, id, request.Title, request.PhotoBase64, request.Description, request.Instructions)
//		if err != nil {
//			c.serveError(w, http.StatusBadRequest, err)
//			return
//		}
//	}
//}

func (c *Recipes) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		c.serveError(w, http.StatusBadRequest, err)
		return
	}

	err = c.recipes.Delete(ctx, id)
	if err != nil {
		c.serveError(w, http.StatusBadRequest, err)
		return
	}
}

// serveError replies to request with specific code and error.
func (c *Recipes) serveError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)

	var response struct {
		Error string `json:"error"`
	}

	response.Error = err.Error()

	if err = json.NewEncoder(w).Encode(response); err != nil {
		log.Println("failed to write json error response", ErrRecipes.Wrap(err))
	}
}
