package test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"citadel/internal/recipe"
	"citadel/internal/session"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRecipe_Authenticated(t *testing.T) {
	req, err := http.NewRequest("GET", server.URL+"/recipes/"+td.RecipeId, nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var r recipe.Recipe
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&r))
	assert.Equal(t, td.RecipeId, r.Id)
	assert.Equal(t, "Test Recipe", r.Title)
	assert.Len(t, r.Ingredients, 1)
	assert.Equal(t, "Water", r.Ingredients[0].Item)
	assert.Len(t, r.Instructions, 1)
	assert.Equal(t, "Boil water", r.Instructions[0])
}

func TestGetRecipe_Unauthenticated(t *testing.T) {
	req, err := http.NewRequest("GET", server.URL+"/recipes/"+td.RecipeId, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestListRecipes_Authenticated(t *testing.T) {
	req, err := http.NewRequest("GET", server.URL+"/recipes", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var recipes []recipe.Recipe
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&recipes))
	assert.NotEmpty(t, recipes)
}

func TestCreateRecipe_Authenticated(t *testing.T) {
	// 10 mins = 600,000,000,000 nanoseconds
	payload := `{
		"title": "New Recipe",
		"description": "New Description",
		"ingredients": [{"amount": 2, "unit": "each", "item": "Eggs"}],
		"instructions": ["Crack eggs", "Fry eggs"],
		"cook_time": 600000000000,
		"serves": 2,
		"cuisine": "American",
		"category": "Main"
	}`
	req, err := http.NewRequest("POST", server.URL+"/recipes", strings.NewReader(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(t, result["recipe_id"])
}

func TestUpdateRecipe_OwnRecipe(t *testing.T) {
	payload := `{"title": "Updated Title"}`
	req, err := http.NewRequest(
		"PATCH",
		server.URL+"/recipes/"+td.RecipeId,
		strings.NewReader(payload),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify update
	req, _ = http.NewRequest("GET", server.URL+"/recipes/"+td.RecipeId, nil)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})
	resp, _ = http.DefaultClient.Do(req)
	defer resp.Body.Close()
	var r recipe.Recipe
	json.NewDecoder(resp.Body).Decode(&r)
	assert.Equal(t, "Updated Title", r.Title)
}

func TestUpdateRecipe_OtherUser(t *testing.T) {
	payload := `{"title": "Hacked"}`
	req, err := http.NewRequest(
		"PATCH",
		server.URL+"/recipes/"+td.RecipeId,
		strings.NewReader(payload),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(
		&http.Cookie{Name: session.CookieName, Value: td.Admin.Session},
	) // Admin is not the owner

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestDeleteRecipe_OwnRecipe(t *testing.T) {
	// Create a recipe to delete
	payload := `{"title": "Delete Me", "ingredients": [], "instructions": [], "cuisine": "American", "difficulty": "Easy", "category": "Main"}`
	req, _ := http.NewRequest("POST", server.URL+"/recipes", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})
	resp, _ := http.DefaultClient.Do(req)
	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	recipeId := result["recipe_id"]
	resp.Body.Close()

	// Delete it
	req, err := http.NewRequest("DELETE", server.URL+"/recipes/"+recipeId, nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify deletion (should return 500 or 404 depending on implementation, ById returns 500 on error)
	req, _ = http.NewRequest("GET", server.URL+"/recipes/"+recipeId, nil)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})
	resp, _ = http.DefaultClient.Do(req)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestDeleteRecipe_OtherUser(t *testing.T) {
	req, err := http.NewRequest("DELETE", server.URL+"/recipes/"+td.RecipeId, nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.Admin.Session})

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
