package queries

import (
	"context"
	"delivery/internal/core/ports"
	"delivery/internal/pkg/errs"
)

type GetAllCouriersQueryHandler interface {
	Handle(ctx context.Context, query GetAllCouriersQuery) (GetAllCouriersResponse, error)
}

type GetBusyCouriersQueryHandler = GetAllCouriersQueryHandler

var _ GetAllCouriersQueryHandler = (*getAllCouriersQueryHandler)(nil)

type getAllCouriersQueryHandler struct {
	unitOfWork        ports.UnitOfWork
	courierRepository ports.CourierRepository
}

func NewGetAllCouriersQueryHandler(unitOfWork ports.UnitOfWork,
	courierRepository ports.CourierRepository) (GetAllCouriersQueryHandler, error) {
	if unitOfWork == nil {
		return nil, errs.NewValueIsRequired("unitOfWork")
	}
	if courierRepository == nil {
		return nil, errs.NewValueIsRequired("courierRepository")
	}
	return &getAllCouriersQueryHandler{unitOfWork: unitOfWork, courierRepository: courierRepository}, nil
}

func NewGetBusyCouriersQueryHandler(unitOfWork ports.UnitOfWork,
	courierRepository ports.CourierRepository) (GetBusyCouriersQueryHandler, error) {
	return NewGetAllCouriersQueryHandler(unitOfWork, courierRepository)
}

func (h *getAllCouriersQueryHandler) Handle(ctx context.Context,
	_ GetAllCouriersQuery) (GetAllCouriersResponse, error) {
	response := GetAllCouriersResponse{Couriers: make([]CourierDTO, 0)}
	err := h.unitOfWork.Do(ctx, func(tx ports.Tx) error {
		couriers, err := h.courierRepository.GetAll(ctx, tx)
		if err != nil {
			return err
		}
		for _, aggregate := range couriers {
			response.Couriers = append(response.Couriers, CourierDTO{
				ID:       aggregate.ID(),
				Name:     aggregate.Name(),
				Location: aggregate.Location(),
			})
		}
		return nil
	})
	if err != nil {
		return GetAllCouriersResponse{}, err
	}
	return response, nil
}
