package commands

import (
	"context"
	"delivery/internal/core/ports"
	"delivery/internal/pkg/errs"
)

type CompleteOrderCommandHandler interface {
	Handle(ctx context.Context, command CompleteOrderCommand) error
}

var _ CompleteOrderCommandHandler = (*completeOrderCommandHandler)(nil)

type completeOrderCommandHandler struct {
	unitOfWork        ports.UnitOfWork
	orderRepository   ports.OrderRepository
	courierRepository ports.CourierRepository
}

func NewCompleteOrderCommandHandler(unitOfWork ports.UnitOfWork, orderRepository ports.OrderRepository, courierRepository ports.CourierRepository) (CompleteOrderCommandHandler, error) {
	if unitOfWork == nil {
		return nil, errs.NewValueIsRequired("unitOfWork")
	}
	if orderRepository == nil {
		return nil, errs.NewValueIsRequired("orderRepository")
	}
	if courierRepository == nil {
		return nil, errs.NewValueIsRequired("courierRepository")
	}
	return &completeOrderCommandHandler{unitOfWork: unitOfWork, orderRepository: orderRepository, courierRepository: courierRepository}, nil
}

func (h *completeOrderCommandHandler) Handle(ctx context.Context, command CompleteOrderCommand) error {
	return h.unitOfWork.Do(ctx, func(tx ports.Tx) error {
		deliveryOrder, err := h.orderRepository.Get(ctx, tx, command.OrderID)
		if err != nil {
			return err
		}
		aggregate, err := h.courierRepository.Get(ctx, tx, command.CourierID)
		if err != nil {
			return err
		}
		if err = aggregate.CompleteOrder(deliveryOrder); err != nil {
			return err
		}
		if err = h.orderRepository.Update(ctx, tx, deliveryOrder); err != nil {
			return err
		}
		return h.courierRepository.Update(ctx, tx, aggregate)
	})
}
