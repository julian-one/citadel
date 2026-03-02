package pokemon

import "github.com/mtslzr/pokeapi-go/structs"

type Pokemon struct {
	Id     int    `db:"pokemon_id"`
	Name   string `db:"name"`
	Height int    `db:"height"`
	Weight int    `db:"weight"`
}

func ToPokemon(response structs.Pokemon) Pokemon {
	return Pokemon{
		Id:     response.ID,
		Name:   response.Name,
		Weight: response.Weight,
		Height: response.Height,
	}
}
