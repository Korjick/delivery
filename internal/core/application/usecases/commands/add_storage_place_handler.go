package commands

import (
	"context"
	"delivery/internal/core/ports"
	"delivery/internal/pkg/errs"
)

type AddStoragePlaceCommandHandler interface {
	Handle(ctx context.Context, command AddStoragePlaceCommand) error
}

type addStoragePlaceCommandHandler struct {
	unitOfWork        ports.UnitOfWork
	courierRepository ports.CourierRepository
}

func NewAddStoragePlaceCommandHandler(unitOfWork ports.UnitOfWork,
	courierRepository ports.CourierRepository) (AddStoragePlaceCommandHandler, error) {
	if unitOfWork == nil {
		return nil, errs.NewValueIsRequired("unitOfWork")
	}
	if courierRepository == nil {
		return nil, errs.NewValueIsRequired("courierRepository")
	}
	return &addStoragePlaceCommandHandler{unitOfWork: unitOfWork, courierRepository: courierRepository}, nil
}

func (h *addStoragePlaceCommandHandler) Handle(ctx context.Context, command AddStoragePlaceCommand) error {
	return h.unitOfWork.Do(ctx, func(tx ports.Tx) error {
		aggregate, err := h.courierRepository.Get(ctx, tx, command.CourierID)
		if err != nil {
			return err
		}
		if err = aggregate.AddStoragePlace(command.Name, command.TotalVolume); err != nil {
			return err
		}
		return h.courierRepository.Update(ctx, tx, aggregate)
	})
}
