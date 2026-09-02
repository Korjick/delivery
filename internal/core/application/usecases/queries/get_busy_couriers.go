package queries

import (
	"delivery/internal/core/domain/kernel"

	"github.com/google/uuid"
)

type GetAllCouriersQuery struct{}

type CourierDTO struct {
	ID       uuid.UUID
	Name     string
	Location kernel.Location
}

type GetAllCouriersResponse struct {
	Couriers []CourierDTO
}
