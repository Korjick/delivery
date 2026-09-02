package commands

import "github.com/google/uuid"

// ProcessBasketConfirmedCommand is the application representation of a
// BasketConfirmed integration event. MessageID identifies a concrete Kafka
// record and is used by Inbox to make processing idempotent.
type ProcessBasketConfirmedCommand struct {
	MessageID string
	OrderID   uuid.UUID
	Street    string
	Volume    int
}
