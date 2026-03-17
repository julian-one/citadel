package pokemon

type Pokemon struct {
	Id     int    `db:"pokemon_id"`
	Name   string `db:"name"`
	Height int    `db:"height"`
	Weight int    `db:"weight"`
}
