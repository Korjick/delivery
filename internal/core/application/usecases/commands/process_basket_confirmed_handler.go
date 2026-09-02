package commands

import (
	"context"
	"delivery/internal/core/domain/model/order"
	"delivery/internal/core/ports"
	"delivery/internal/pkg/errs"
	"strings"
)

// ProcessBasketConfirmedCommandHandler creates an order from a Basket event
// exactly once. The Inbox record and the order are persisted in one Unit of
// Work transaction, so a consumer retry after a successful transaction cannot
// create a second order.
type ProcessBasketConfirmedCommandHandler interface {
	Handle(ctx context.Context, command ProcessBasketConfirmedCommand) error
}

type processBasketConfirmedCommandHandler struct {
	unitOfWork      ports.UnitOfWork
	inboxRepository ports.InboxRepository
	orderRepository ports.OrderRepository
	geoClient       ports.GeoClient
}

func NewProcessBasketConfirmedCommandHandler(
	unitOfWork ports.UnitOfWork,
	inboxRepository ports.InboxRepository,
	orderRepository ports.OrderRepository,
	geoClient ports.GeoClient,
) (ProcessBasketConfirmedCommandHandler, error) {
	if unitOfWork == nil {
		return nil, errs.NewValueIsRequired("unitOfWork")
	}
	if inboxRepository == nil {
		return nil, errs.NewValueIsRequired("inboxRepository")
	}
	if orderRepository == nil {
		return nil, errs.NewValueIsRequired("orderRepository")
	}
	if geoClient == nil {
		return nil, errs.NewValueIsRequired("geoClient")
	}

	return &processBasketConfirmedCommandHandler{
		unitOfWork:      unitOfWork,
		inboxRepository: inboxRepository,
		orderRepository: orderRepository,
		geoClient:       geoClient,
	}, nil
}

func (h *processBasketConfirmedCommandHandler) Handle(ctx context.Context, command ProcessBasketConfirmedCommand) error {
	if strings.TrimSpace(command.MessageID) == "" {
		return errs.NewValueIsRequired("messageID")
	}

	// Geo is an external call. Resolve it before opening a database transaction
	// so the transaction holds no database locks while waiting for Geo.
	location, err := h.geoClient.GetLocation(ctx, command.Street)
	if err != nil {
		return err
	}

	return h.unitOfWork.Do(ctx, func(tx ports.Tx) error {
		isNew, err := h.inboxRepository.TryAdd(ctx, tx, command.MessageID)
		if err != nil {
			return err
		}
		if !isNew {
			return nil
		}

		aggregate, err := order.NewOrder(command.OrderID, location, command.Volume)
		if err != nil {
			return err
		}
		return h.orderRepository.Add(ctx, tx, aggregate)
	})
}
