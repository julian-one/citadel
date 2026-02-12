package database

type Order string

const (
	Descending Order = "DESC"
	Ascending  Order = "ASC"
)

func (o Order) Valid() bool {
	switch o {
	case Descending, Ascending:
		return true
	default:
		return false
	}
}

// OrderBy represents a single column ordering for database queries
type OrderBy struct {
	Column string
	Order  Order
}
