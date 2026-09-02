package commands

import (
	"context"
	"delivery/internal/core/ports"
	"delivery/internal/pkg/errs"
)

type MoveCouriersCommandHandler interface {
	Handle(ctx context.Context, command MoveCouriersCommand) error
}

type moveCouriersCommandHandler struct {
	unitOfWork        ports.UnitOfWork
	orderRepository   ports.OrderRepository
	courierRepository ports.CourierRepository
}

func NewMoveCouriersCommandHandler(unitOfWork ports.UnitOfWork, orderRepository ports.OrderRepository, courierRepository ports.CourierRepository) (MoveCouriersCommandHandler, error) {
	if unitOfWork == nil {
		return nil, errs.NewValueIsRequired("unitOfWork")
	}
	if orderRepository == nil {
		return nil, errs.NewValueIsRequired("orderRepository")
	}
	if courierRepository == nil {
		return nil, errs.NewValueIsRequired("courierRepository")
	}
	return &moveCouriersCommandHandler{unitOfWork: unitOfWork, orderRepository: orderRepository, courierRepository: courierRepository}, nil
}

func (h *moveCouriersCommandHandler) Handle(ctx context.Context, _ MoveCouriersCommand) error {
	return h.unitOfWork.Do(ctx, func(tx ports.Tx) error {
		assignedOrders, err := h.orderRepository.GetAllAssigned(ctx, tx)
		if err != nil {
			return err
		}
		for _, deliveryOrder := range assignedOrders {
			courierID := deliveryOrder.CourierID()
			if courierID == nil {
				continue
			}
			assignedCourier, err := h.courierRepository.Get(ctx, tx, *courierID)
			if err != nil {
				return err
			}
			if err = assignedCourier.Move(deliveryOrder.Location()); err != nil {
				return err
			}
			courierLocation := assignedCourier.Location()
			orderLocation := deliveryOrder.Location()
			if courierLocation.Equals(&orderLocation) {
				if err = assignedCourier.CompleteOrder(deliveryOrder); err != nil {
					return err
				}
			}
			if err = h.courierRepository.Update(ctx, tx, assignedCourier); err != nil {
				return err
			}
			if err = h.orderRepository.Update(ctx, tx, deliveryOrder); err != nil {
				return err
			}
		}
		return nil
	})
}
