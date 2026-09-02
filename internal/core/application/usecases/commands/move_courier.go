package commands

import (
	"delivery/internal/core/domain/kernel"

	"github.com/google/uuid"
)

type MoveCourierCommand struct {
	CourierID uuid.UUID
	Target    kernel.Location
}
