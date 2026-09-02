package queries

import (
	"context"
	"delivery/internal/core/ports"
	"delivery/internal/pkg/errs"
)

type GetNotCompletedOrdersQueryHandler interface {
	Handle(ctx context.Context, query GetNotCompletedOrdersQuery) (GetNotCompletedOrdersResponse, error)
}

type getNotCompletedOrdersQueryHandler struct {
	unitOfWork      ports.UnitOfWork
	orderRepository ports.OrderRepository
}

func NewGetNotCompletedOrdersQueryHandler(unitOfWork ports.UnitOfWork,
	orderRepository ports.OrderRepository) (GetNotCompletedOrdersQueryHandler, error) {
	if unitOfWork == nil {
		return nil, errs.NewValueIsRequired("unitOfWork")
	}
	if orderRepository == nil {
		return nil, errs.NewValueIsRequired("orderRepository")
	}
	return &getNotCompletedOrdersQueryHandler{unitOfWork: unitOfWork, orderRepository: orderRepository}, nil
}

func (h *getNotCompletedOrdersQueryHandler) Handle(ctx context.Context,
	_ GetNotCompletedOrdersQuery) (GetNotCompletedOrdersResponse, error) {
	response := GetNotCompletedOrdersResponse{Orders: make([]OrderDTO, 0)}
	err := h.unitOfWork.Do(ctx, func(tx ports.Tx) error {
		orders, err := h.orderRepository.GetAllNotCompleted(ctx, tx)
		if err != nil {
			return err
		}
		for _, aggregate := range orders {
			response.Orders = append(response.Orders, OrderDTO{ID: aggregate.ID(), Location: aggregate.Location()})
		}
		return nil
	})
	if err != nil {
		return GetNotCompletedOrdersResponse{}, err
	}
	return response, nil
}
