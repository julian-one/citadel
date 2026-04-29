package pokemon

type Pokemon struct {
	ID     int    `db:"pokemon_id"`
	Name   string `db:"name"`
	Height int    `db:"height"`
	Weight int    `db:"weight"`
}
