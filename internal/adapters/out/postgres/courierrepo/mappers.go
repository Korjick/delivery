package courierrepo

import (
	"delivery/internal/core/domain/kernel"
	"delivery/internal/core/domain/model/courier"
)

func DomainToDTO(aggregate *courier.Courier) CourierDTO {
	dto := CourierDTO{
		ID:            aggregate.ID(),
		Name:          aggregate.Name(),
		Speed:         aggregate.Speed(),
		LocationX:     aggregate.Location().X(),
		LocationY:     aggregate.Location().Y(),
		StoragePlaces: make([]*StoragePlaceDTO, 0, len(aggregate.StoragePlaces())),
	}
	for _, storagePlace := range aggregate.StoragePlaces() {
		dto.StoragePlaces = append(dto.StoragePlaces, &StoragePlaceDTO{
			ID:          storagePlace.ID(),
			CourierID:   aggregate.ID(),
			Name:        storagePlace.Name(),
			TotalVolume: storagePlace.TotalVolume(),
			OrderID:     storagePlace.OrderID(),
		})
	}
	return dto
}

func DTOToDomain(dto CourierDTO) *courier.Courier {
	storagePlaces := make([]*courier.StoragePlace, 0, len(dto.StoragePlaces))
	for _, storagePlaceDTO := range dto.StoragePlaces {
		storagePlaces = append(storagePlaces, courier.RestoreStoragePlace(
			storagePlaceDTO.ID,
			storagePlaceDTO.Name,
			storagePlaceDTO.TotalVolume,
			storagePlaceDTO.OrderID,
		))
	}
	return courier.RestoreCourier(
		dto.ID,
		dto.Name,
		dto.Speed,
		kernel.MustNewLocation(dto.LocationX, dto.LocationY),
		storagePlaces,
	)
}
