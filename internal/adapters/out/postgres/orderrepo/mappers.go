package orderrepo

import (
	"delivery/internal/core/domain/kernel"
	"delivery/internal/core/domain/model/order"
)

func DomainToDTO(aggregate *order.Order) OrderDTO {
	return OrderDTO{
		ID:        aggregate.ID(),
		CourierID: aggregate.CourierID(),
		LocationX: aggregate.Location().X(),
		LocationY: aggregate.Location().Y(),
		Volume:    aggregate.Volume(),
		Status:    aggregate.Status(),
	}
}

func DTOToDomain(dto OrderDTO) *order.Order {
	return order.RestoreOrder(
		dto.ID,
		dto.CourierID,
		kernel.MustNewLocation(dto.LocationX, dto.LocationY),
		dto.Volume,
		dto.Status,
	)
}
