package commands

import "github.com/google/uuid"

type AddStoragePlaceCommand struct {
	CourierID   uuid.UUID
	Name        string
	TotalVolume int
}
