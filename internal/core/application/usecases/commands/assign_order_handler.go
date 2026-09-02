package commands

import (
	"context"
	"delivery/internal/core/domain/services"
	"delivery/internal/core/ports"
	"delivery/internal/pkg/errs"
	"errors"
)

type AssignOrderCommandHandler interface {
	Handle(ctx context.Context, command AssignOrderCommand) error
}

type assignOrderCommandHandler struct {
	unitOfWork        ports.UnitOfWork
	orderRepository   ports.OrderRepository
	courierRepository ports.CourierRepository
	dispatcher        services.OrderDispatcher
}

func NewAssignOrderCommandHandler(unitOfWork ports.UnitOfWork, orderRepository ports.OrderRepository,
	courierRepository ports.CourierRepository, dispatcher services.OrderDispatcher) (AssignOrderCommandHandler, error) {
	if unitOfWork == nil {
		return nil, errs.NewValueIsRequired("unitOfWork")
	}
	if orderRepository == nil {
		return nil, errs.NewValueIsRequired("orderRepository")
	}
	if courierRepository == nil {
		return nil, errs.NewValueIsRequired("courierRepository")
	}
	if dispatcher == nil {
		return nil, errs.NewValueIsRequired("dispatcher")
	}
	return &assignOrderCommandHandler{unitOfWork: unitOfWork, orderRepository: orderRepository,
		courierRepository: courierRepository, dispatcher: dispatcher}, nil
}

func (h *assignOrderCommandHandler) Handle(ctx context.Context, _ AssignOrderCommand) error {
	return h.unitOfWork.Do(ctx, func(tx ports.Tx) error {
		createdOrders, err := h.orderRepository.GetAllCreated(ctx, tx)
		if err != nil {
			return err
		}
		if len(createdOrders) == 0 {
			return nil
		}
		freeCouriers, err := h.courierRepository.GetAllFree(ctx, tx)
		if err != nil {
			return err
		}
		if len(freeCouriers) == 0 {
			return nil
		}

		for _, deliveryOrder := range createdOrders {
			winner, err := h.dispatcher.Dispatch(deliveryOrder, freeCouriers)
			if errors.Is(err, services.ErrNoSuitableCourierFound) {
				continue
			}
			if err != nil {
				return err
			}
			if err = h.orderRepository.Update(ctx, tx, deliveryOrder); err != nil {
				return err
			}
			return h.courierRepository.Update(ctx, tx, winner)
		}
		return nil
	})
}
