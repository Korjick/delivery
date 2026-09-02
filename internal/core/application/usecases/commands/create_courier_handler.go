package commands

import (
	"context"
	"delivery/internal/core/domain/kernel"
	"delivery/internal/core/domain/model/courier"
	"delivery/internal/core/ports"
	"delivery/internal/pkg/errs"

	"github.com/google/uuid"
)

type CreateCourierCommandHandler interface {
	Handle(ctx context.Context, command CreateCourierCommand) (uuid.UUID, error)
}

type createCourierCommandHandler struct {
	unitOfWork        ports.UnitOfWork
	courierRepository ports.CourierRepository
}

func NewCreateCourierCommandHandler(unitOfWork ports.UnitOfWork,
	courierRepository ports.CourierRepository) (CreateCourierCommandHandler, error) {
	if unitOfWork == nil {
		return nil, errs.NewValueIsRequired("unitOfWork")
	}
	if courierRepository == nil {
		return nil, errs.NewValueIsRequired("courierRepository")
	}
	return &createCourierCommandHandler{unitOfWork: unitOfWork, courierRepository: courierRepository}, nil
}

func (h *createCourierCommandHandler) Handle(ctx context.Context, command CreateCourierCommand) (uuid.UUID, error) {
	var courierID uuid.UUID
	err := h.unitOfWork.Do(ctx, func(tx ports.Tx) error {
		aggregate, err := courier.NewCourier(command.Name, command.Speed, *kernel.RandomLocation())
		if err != nil {
			return err
		}
		if err = aggregate.AddStoragePlace("Рюкзак", 20); err != nil {
			return err
		}
		if err = h.courierRepository.Add(ctx, tx, aggregate); err != nil {
			return err
		}
		courierID = aggregate.ID()
		return nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	return courierID, nil
}
