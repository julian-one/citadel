package database

import (
	"fmt"
	"strings"
)

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

// ParseOrderBy parses a comma-separated order_by string (e.g. "column:asc,column:desc")
// into a slice of OrderBy values.
func ParseOrderBy(raw string) ([]OrderBy, error) {
	var result []OrderBy

	// Split the raw string by commas to get individual order specifications
	orders := strings.SplitSeq(raw, ",")

	// Loop through each order specification and parse it into an OrderBy struct
	for order := range orders {

		// Split each order by the colon to separate the column and the order direction
		parts := strings.Split(strings.TrimSpace(order), ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf(
				"invalid order_by format: %s (expected 'column:order')",
				order,
			)
		}

		// NOTE: The caller is responsible for validating that the column name is valid and exists in the database schema.
		column := parts[0]

		// Convert the order direction to uppercase and validate it
		orderStr := strings.ToUpper(parts[1])
		orderDir := Order(orderStr)
		if !orderDir.Valid() {
			return nil, fmt.Errorf(
				"invalid order direction: %s (must be 'asc' or 'desc')",
				parts[1],
			)
		}

		result = append(result, OrderBy{
			Column: column,
			Order:  orderDir,
		})
	}

	return result, nil
}
