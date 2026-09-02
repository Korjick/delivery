package commands

import (
	"context"
	"delivery/internal/core/ports"
	"delivery/internal/pkg/errs"
)

type MoveCourierCommandHandler interface {
	Handle(ctx context.Context, command MoveCourierCommand) error
}

var _ MoveCourierCommandHandler = (*moveCourierCommandHandler)(nil)

type moveCourierCommandHandler struct {
	unitOfWork        ports.UnitOfWork
	courierRepository ports.CourierRepository
}

func NewMoveCourierCommandHandler(unitOfWork ports.UnitOfWork, courierRepository ports.CourierRepository) (MoveCourierCommandHandler, error) {
	if unitOfWork == nil {
		return nil, errs.NewValueIsRequired("unitOfWork")
	}
	if courierRepository == nil {
		return nil, errs.NewValueIsRequired("courierRepository")
	}
	return &moveCourierCommandHandler{unitOfWork: unitOfWork, courierRepository: courierRepository}, nil
}

func (h *moveCourierCommandHandler) Handle(ctx context.Context, command MoveCourierCommand) error {
	return h.unitOfWork.Do(ctx, func(tx ports.Tx) error {
		aggregate, err := h.courierRepository.Get(ctx, tx, command.CourierID)
		if err != nil {
			return err
		}
		if err = aggregate.Move(command.Target); err != nil {
			return err
		}
		return h.courierRepository.Update(ctx, tx, aggregate)
	})
}
