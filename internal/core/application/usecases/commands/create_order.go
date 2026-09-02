package commands

import "github.com/google/uuid"

type CreateOrderCommand struct {
	OrderID uuid.UUID
	Street  string
	Volume  int
}
