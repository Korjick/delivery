package queries

import (
	"delivery/internal/core/domain/kernel"

	"github.com/google/uuid"
)

type GetNotCompletedOrdersQuery struct{}

type OrderDTO struct {
	ID       uuid.UUID
	Location kernel.Location
}

type GetNotCompletedOrdersResponse struct {
	Orders []OrderDTO
}
