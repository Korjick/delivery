package commands

import "github.com/google/uuid"

type CompleteOrderCommand struct {
	CourierID uuid.UUID
	OrderID   uuid.UUID
}
