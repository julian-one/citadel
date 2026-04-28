package pokemon

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/mtslzr/pokeapi-go/structs"
)

func Create(ctx context.Context, db sqlx.ExtContext, resp structs.Pokemon) (*Pokemon, error) {
	var p Pokemon
	err := db.QueryRowxContext(
		ctx,
		`INSERT INTO pokemon (pokemon_id, name, height, weight) 
		 	VALUES (?, ?, ?, ?) RETURNING pokemon_id, name, height, weight`,
		resp.ID, resp.Name, resp.Height, resp.Weight,
	).StructScan(&p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
