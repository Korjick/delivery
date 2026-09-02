package queries

import (
	"context"
	"delivery/internal/core/domain/kernel"
	"delivery/internal/core/domain/model/courier"
	"delivery/internal/core/domain/model/order"
	"delivery/internal/core/ports"
	"delivery/internal/pkg/ddd"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestGetAllCouriersQueryHandler_Handle(t *testing.T) {
	aggregate, err := courier.NewCourier("Alex", 1, kernel.MustNewLocation(1, 1))
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewGetAllCouriersQueryHandler(queryUnitOfWork{}, &queryCourierRepository{busy: []*courier.Courier{aggregate}})
	if err != nil {
		t.Fatal(err)
	}
	response, err := handler.Handle(context.Background(), GetAllCouriersQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Couriers) != 1 || response.Couriers[0].ID != aggregate.ID() || response.Couriers[0].Name != "Alex" || response.Couriers[0].Location != aggregate.Location() {
		t.Errorf("Handle() = %#v, want courier DTO", response)
	}
}

func TestGetNotCompletedOrdersQueryHandler_Handle(t *testing.T) {
	first, err := order.NewOrder(uuid.New(), kernel.MustNewLocation(1, 1), 1)
	if err != nil {
		t.Fatal(err)
	}
	second, err := order.NewOrder(uuid.New(), kernel.MustNewLocation(2, 2), 1)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewGetNotCompletedOrdersQueryHandler(queryUnitOfWork{}, &queryOrderRepository{notCompleted: []*order.Order{first, second}})
	if err != nil {
		t.Fatal(err)
	}
	response, err := handler.Handle(context.Background(), GetNotCompletedOrdersQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Orders) != 2 || response.Orders[0].ID != first.ID() || response.Orders[1].Location != second.Location() {
		t.Errorf("Handle() = %#v, want order DTOs", response)
	}
}

type queryUnitOfWork struct{}

func (queryUnitOfWork) Do(_ context.Context, fn func(tx ports.Tx) error) error { return fn(&queryTx{}) }

type queryTx struct{}

func (*queryTx) DB() *gorm.DB              { return nil }
func (*queryTx) AddEvent(ddd.DomainEvent)  {}
func (*queryTx) Events() []ddd.DomainEvent { return nil }

type queryCourierRepository struct{ busy []*courier.Courier }

func (*queryCourierRepository) Add(context.Context, ports.Tx, *courier.Courier) error    { return nil }
func (*queryCourierRepository) Update(context.Context, ports.Tx, *courier.Courier) error { return nil }
func (*queryCourierRepository) Get(context.Context, ports.Tx, uuid.UUID) (*courier.Courier, error) {
	return nil, nil
}
func (r *queryCourierRepository) GetAll(context.Context, ports.Tx) ([]*courier.Courier, error) {
	return r.busy, nil
}
func (*queryCourierRepository) GetAllFree(context.Context, ports.Tx) ([]*courier.Courier, error) {
	return nil, nil
}

type queryOrderRepository struct{ notCompleted []*order.Order }

func (*queryOrderRepository) Add(context.Context, ports.Tx, *order.Order) error    { return nil }
func (*queryOrderRepository) Update(context.Context, ports.Tx, *order.Order) error { return nil }
func (*queryOrderRepository) Get(context.Context, ports.Tx, uuid.UUID) (*order.Order, error) {
	return nil, nil
}
func (*queryOrderRepository) GetAllCreated(context.Context, ports.Tx) ([]*order.Order, error) {
	return nil, nil
}
func (*queryOrderRepository) GetAllAssigned(context.Context, ports.Tx) ([]*order.Order, error) {
	return nil, nil
}
func (r *queryOrderRepository) GetAllNotCompleted(context.Context, ports.Tx) ([]*order.Order, error) {
	return r.notCompleted, nil
}
