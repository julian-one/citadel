package test

import (
	"encoding/json"
	"net/http"
	"testing"

	"citadel/internal/pokemon"
	"citadel/internal/session"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchPokemon_Unauthenticated(t *testing.T) {
	req, err := http.NewRequest("GET", server.URL+"/pokemon?name=pikachu", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestSearchPokemon_MissingName(t *testing.T) {
	req, err := http.NewRequest("GET", server.URL+"/pokemon", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, "name query parameter is required", result["error"])
}

func TestSearchPokemon_EmptyName(t *testing.T) {
	req, err := http.NewRequest("GET", server.URL+"/pokemon?name=", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSearchPokemon_FromDatabase(t *testing.T) {
	req, err := http.NewRequest("GET", server.URL+"/pokemon?name=bulbasaur", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var p pokemon.Pokemon
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&p))
	assert.Equal(t, 1, p.Id)
	assert.Equal(t, "bulbasaur", p.Name)
	assert.Equal(t, 7, p.Height)
	assert.Equal(t, 69, p.Weight)
}

func TestSearchPokemon_CaseInsensitive(t *testing.T) {
	req, err := http.NewRequest("GET", server.URL+"/pokemon?name=Bulbasaur", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var p pokemon.Pokemon
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&p))
	assert.Equal(t, "bulbasaur", p.Name)
}

func TestSearchPokemon_FromPokeAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping external API test in short mode")
	}

	req, err := http.NewRequest("GET", server.URL+"/pokemon?name=charmander", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var p pokemon.Pokemon
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&p))
	assert.Equal(t, "charmander", p.Name)
	assert.Equal(t, 4, p.Id)

	// Searching again should return from the database (cached)
	req, err = http.NewRequest("GET", server.URL+"/pokemon?name=charmander", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp2, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	var p2 pokemon.Pokemon
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&p2))
	assert.Equal(t, "charmander", p2.Name)
}

func TestSearchPokemon_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping external API test in short mode")
	}

	req, err := http.NewRequest("GET", server.URL+"/pokemon?name=notarealpokemon", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: session.CookieName, Value: td.User.Session})

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var result map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, "pokemon not found", result["error"])
}
