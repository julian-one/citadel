package route

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"citadel/internal/pokemon"

	"github.com/jmoiron/sqlx"
	"github.com/mtslzr/pokeapi-go"
)

func SearchPokemon(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		name := r.URL.Query().Get("name")
		if name == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).
				Encode(map[string]string{"error": "name query parameter is required"})
			return
		}
		name = strings.ToLower(name)
		name = strings.ReplaceAll(name, " ", "-")

		// check if pokemon exists in the database
		var p pokemon.Pokemon
		exists := pokemon.Exists(ctx, db, name)
		if exists {
			fetched, err := pokemon.ByName(ctx, db, name)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).
					Encode(map[string]string{"error": "failed to fetch pokemon data from database"})
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(fetched)
			return
		}

		// fetch from the pokeapi
		logger.Info("fetching from pokeapi", "name", name)
		response, err := pokeapi.Pokemon(name)
		if err != nil {
			if strings.Contains(
				err.Error(),
				"invalid character 'N' looking for beginning of value",
			) {
				w.WriteHeader(http.StatusNotFound)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).
					Encode(map[string]string{"error": "pokemon not found"})
				return
			}
			logger.Error("failed to fetch pokemon", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).
				Encode(map[string]string{"error": "failed to fetch pokemon data"})
			return
		}

		// save to the database
		created, err := pokemon.Create(ctx, db, response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).
				Encode(map[string]string{"error": "failed to save pokemon data"})
			return
		}
		if created != nil {
			p = *created
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
	}
}
