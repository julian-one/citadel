package route

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/mtslzr/pokeapi-go"
	"github.com/mtslzr/pokeapi-go/structs"
)

type Pokemon struct {
	Id     int    `db:"pokemon_id"`
	Name   string `db:"name"`
	Height int    `db:"height"`
	Weight int    `db:"weight"`
}

func toPokemon(response structs.Pokemon) Pokemon {
	return Pokemon{
		Id:     response.ID,
		Name:   response.Name,
		Weight: response.Weight,
		Height: response.Height,
	}
}

func exists(ctx context.Context, db *sqlx.DB, name string) bool {
	var exists bool
	err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pokemon WHERE name = ?)`, name).
		Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

func save(ctx context.Context, db *sqlx.DB, p Pokemon) error {
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO pokemon (pokemon_id, name, height, weight) VALUES (?, ?, ?, ?)`,
		p.Id, p.Name, p.Height, p.Weight)
	if err != nil {
		return err
	}
	return nil
}

func fetch(ctx context.Context, db *sqlx.DB, name string) (*Pokemon, error) {
	var p Pokemon
	err := db.GetContext(
		ctx,
		&p,
		`SELECT pokemon_id, name, height, weight FROM pokemon WHERE name = ?`,
		name,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func GetPokemon(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// get the name from a query parameter
		name := r.URL.Query().Get("name")
		if name == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).
				Encode(map[string]string{"error": "name query parameter is required"})
			return
		}

		// case insensitive name
		name = strings.ToLower(name)
		name = strings.ReplaceAll(name, " ", "-")
		logger.Info("fetching pokemon", slog.String("name", name))

		// check if pokemon exists in the database
		var p Pokemon
		exists := exists(ctx, db, name)
		if exists {
			logger.Info("pokemon exists in database", slog.String("name", name))

			fetched, err := fetch(ctx, db, name)
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
		logger.Info("fetching from pokeapi", slog.String("name", name))
		response, err := pokeapi.Pokemon(name)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character 'N' looking for beginning of value") {
				w.WriteHeader(http.StatusNotFound)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).
					Encode(map[string]string{"error": "pokemon not found"})
				return
			}
			logger.Error("failed to fetch pokemon", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).
				Encode(map[string]string{"error": "failed to fetch pokemon data"})
			return
		}

		// convert to our Pokemon struct and save to the database
		p = toPokemon(response)
		err = save(ctx, db, p)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).
				Encode(map[string]string{"error": "failed to save pokemon data"})
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
	}
}
