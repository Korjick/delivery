package http

import (
	"delivery/internal/core/application/usecases/commands"
	"delivery/internal/core/application/usecases/queries"
	"delivery/internal/core/domain/kernel"
	"delivery/internal/generated/servers"
	"delivery/internal/pkg/errs"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type Handler struct {
	createOrder     commands.CreateOrderCommandHandler
	createCourier   commands.CreateCourierCommandHandler
	moveCourier     commands.MoveCourierCommandHandler
	completeOrder   commands.CompleteOrderCommandHandler
	getAllCouriers  queries.GetAllCouriersQueryHandler
	getActiveOrders queries.GetNotCompletedOrdersQueryHandler
}

func NewHandler(
	createOrder commands.CreateOrderCommandHandler,
	createCourier commands.CreateCourierCommandHandler,
	moveCourier commands.MoveCourierCommandHandler,
	completeOrder commands.CompleteOrderCommandHandler,
	getAllCouriers queries.GetAllCouriersQueryHandler,
	getActiveOrders queries.GetNotCompletedOrdersQueryHandler,
) (*Handler, error) {
	if createOrder == nil || createCourier == nil || moveCourier == nil || completeOrder == nil ||
		getAllCouriers == nil || getActiveOrders == nil {
		return nil, errs.NewValueIsRequired("handler dependency")
	}
	return &Handler{
		createOrder:     createOrder,
		createCourier:   createCourier,
		moveCourier:     moveCourier,
		completeOrder:   completeOrder,
		getAllCouriers:  getAllCouriers,
		getActiveOrders: getActiveOrders,
	}, nil
}

func (h *Handler) GetCouriers(ctx echo.Context) error {
	response, err := h.getAllCouriers.Handle(ctx.Request().Context(), queries.GetAllCouriersQuery{})
	if err != nil {
		return writeError(ctx, err)
	}
	couriers := make([]servers.Courier, 0, len(response.Couriers))
	for _, courier := range response.Couriers {
		couriers = append(couriers, servers.Courier{
			Id:       courier.ID,
			Name:     courier.Name,
			Location: servers.Location{X: courier.Location.X(), Y: courier.Location.Y()},
		})
	}
	return ctx.JSON(http.StatusOK, couriers)
}

func (h *Handler) CreateCourier(ctx echo.Context) error {
	var request servers.NewCourier
	if err := ctx.Bind(&request); err != nil {
		return writeError(ctx, err)
	}
	// The current HTTP contract has no speed field, so a walking speed is used.
	courierID, err := h.createCourier.Handle(ctx.Request().Context(), commands.CreateCourierCommand{Name: request.Name, Speed: 1})
	if err != nil {
		return writeError(ctx, err)
	}
	return ctx.JSON(http.StatusCreated, servers.CreateCourierResponse{CourierId: courierID})
}

func (h *Handler) MoveCourier(ctx echo.Context, courierID openapi_types.UUID) error {
	var request servers.Location
	if err := ctx.Bind(&request); err != nil {
		return writeError(ctx, err)
	}
	err := h.moveCourier.Handle(ctx.Request().Context(), commands.MoveCourierCommand{
		CourierID: courierID,
		Target:    kernelLocation(request),
	})
	if err != nil {
		return writeError(ctx, err)
	}
	return ctx.NoContent(http.StatusOK)
}

func (h *Handler) CompleteOrder(ctx echo.Context, courierID openapi_types.UUID, orderID openapi_types.UUID) error {
	err := h.completeOrder.Handle(ctx.Request().Context(), commands.CompleteOrderCommand{
		CourierID: courierID,
		OrderID:   orderID,
	})
	if err != nil {
		return writeError(ctx, err)
	}
	return ctx.NoContent(http.StatusOK)
}

func (h *Handler) CreateOrder(ctx echo.Context) error {
	var request servers.NewOrder
	if err := ctx.Bind(&request); err != nil {
		return writeError(ctx, err)
	}
	err := h.createOrder.Handle(ctx.Request().Context(), commands.CreateOrderCommand{
		OrderID: request.Id,
		Street:  request.Address.Street,
		Volume:  request.Volume,
	})
	if err != nil {
		return writeError(ctx, err)
	}
	return ctx.JSON(http.StatusCreated, servers.CreateOrderResponse{OrderId: request.Id})
}

func (h *Handler) GetOrders(ctx echo.Context) error {
	response, err := h.getActiveOrders.Handle(ctx.Request().Context(), queries.GetNotCompletedOrdersQuery{})
	if err != nil {
		return writeError(ctx, err)
	}
	orders := make([]servers.Order, 0, len(response.Orders))
	for _, deliveryOrder := range response.Orders {
		orders = append(orders, servers.Order{
			Id:       deliveryOrder.ID,
			Location: servers.Location{X: deliveryOrder.Location.X(), Y: deliveryOrder.Location.Y()},
		})
	}
	return ctx.JSON(http.StatusOK, orders)
}

func kernelLocation(location servers.Location) kernel.Location {
	result, err := kernel.NewLocation(location.X, location.Y)
	if err != nil {
		return kernel.Location{}
	}
	return *result
}

func writeError(ctx echo.Context, err error) error {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, errs.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, errs.ErrValueIsRequired),
		errors.Is(err, errs.ErrValueIsInvalid),
		errors.Is(err, errs.ErrValueMustBeGreaterOrEqual),
		errors.Is(err, errs.ErrValueMustBeGreaterThan),
		errors.Is(err, errs.ErrValueMustBeLessOrEqual),
		errors.Is(err, errs.ErrValueMustBeLessThan),
		errors.Is(err, errs.ErrMustBeBetween):
		status = http.StatusBadRequest
	}
	return ctx.JSON(status, servers.Error{Code: int32(status), Message: err.Error()})
}
