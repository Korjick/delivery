package commands

import (
	"context"
	"delivery/internal/core/domain/model/order"
	"delivery/internal/core/ports"
	"delivery/internal/pkg/errs"
)

type CreateOrderCommandHandler interface {
	Handle(ctx context.Context, command CreateOrderCommand) error
}

type createOrderCommandHandler struct {
	unitOfWork      ports.UnitOfWork
	orderRepository ports.OrderRepository
	geoClient       ports.GeoClient
}

func NewCreateOrderCommandHandler(unitOfWork ports.UnitOfWork,
	orderRepository ports.OrderRepository,
	geoClient ports.GeoClient) (CreateOrderCommandHandler, error) {
	if unitOfWork == nil {
		return nil, errs.NewValueIsRequired("unitOfWork")
	}
	if orderRepository == nil {
		return nil, errs.NewValueIsRequired("orderRepository")
	}
	if geoClient == nil {
		return nil, errs.NewValueIsRequired("geoClient")
	}
	return &createOrderCommandHandler{unitOfWork: unitOfWork, orderRepository: orderRepository, geoClient: geoClient}, nil
}

func (h *createOrderCommandHandler) Handle(ctx context.Context, command CreateOrderCommand) error {
	location, err := h.geoClient.GetLocation(ctx, command.Street)
	if err != nil {
		return err
	}

	return h.unitOfWork.Do(ctx, func(tx ports.Tx) error {
		aggregate, err := order.NewOrder(command.OrderID, location, command.Volume)
		if err != nil {
			return err
		}
		return h.orderRepository.Add(ctx, tx, aggregate)
	})
}
