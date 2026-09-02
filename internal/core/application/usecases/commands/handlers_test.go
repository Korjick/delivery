package commands

import (
	"context"
	"delivery/internal/core/domain/kernel"
	"delivery/internal/core/domain/model/courier"
	"delivery/internal/core/domain/model/order"
	"delivery/internal/core/domain/services"
	"delivery/internal/core/ports"
	"delivery/internal/pkg/ddd"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestCreateOrderCommandHandler_Handle(t *testing.T) {
	orderRepository := &fakeOrderRepository{}
	handler, err := NewCreateOrderCommandHandler(fakeUnitOfWork{}, orderRepository, fakeGeoClient{location: kernel.MustNewLocation(3, 4)})
	if err != nil {
		t.Fatal(err)
	}
	orderID := uuid.New()
	if err = handler.Handle(context.Background(), CreateOrderCommand{OrderID: orderID, Street: "Lenina", Volume: 3}); err != nil {
		t.Fatal(err)
	}
	if orderRepository.added == nil || orderRepository.added.ID() != orderID || orderRepository.added.Status() != order.StatusCreated {
		t.Errorf("created order = %#v, want created order %s", orderRepository.added, orderID)
	}
}

type fakeGeoClient struct {
	location kernel.Location
	err      error
}

func (c fakeGeoClient) GetLocation(_ context.Context, _ string) (kernel.Location, error) {
	return c.location, c.err
}

func (fakeGeoClient) Close() error { return nil }

func TestAssignOrderCommandHandler_Handle(t *testing.T) {
	deliveryOrder, err := order.NewOrder(uuid.New(), kernel.MustNewLocation(5, 5), 1)
	if err != nil {
		t.Fatal(err)
	}
	aggregate, err := courier.NewCourier("Alex", 2, kernel.MustNewLocation(1, 1))
	if err != nil {
		t.Fatal(err)
	}
	orderRepository := &fakeOrderRepository{created: deliveryOrder}
	courierRepository := &fakeCourierRepository{free: []*courier.Courier{aggregate}}
	handler, err := NewAssignOrderCommandHandler(fakeUnitOfWork{}, orderRepository, courierRepository, services.NewOrderDispatcher())
	if err != nil {
		t.Fatal(err)
	}
	if err = handler.Handle(context.Background(), AssignOrderCommand{}); err != nil {
		t.Fatal(err)
	}
	if deliveryOrder.Status() != order.StatusAssigned || deliveryOrder.CourierID() == nil || *deliveryOrder.CourierID() != aggregate.ID() {
		t.Errorf("order was not assigned: %#v", deliveryOrder)
	}
	if orderRepository.updated != deliveryOrder || courierRepository.updated != aggregate {
		t.Error("assigned order and courier were not persisted")
	}
}

func TestMoveCouriersCommandHandler_Handle(t *testing.T) {
	deliveryOrder, err := order.NewOrder(uuid.New(), kernel.MustNewLocation(5, 5), 1)
	if err != nil {
		t.Fatal(err)
	}
	aggregate, err := courier.NewCourier("Alex", 1, kernel.MustNewLocation(4, 5))
	if err != nil {
		t.Fatal(err)
	}
	if err = aggregate.TakeOrder(deliveryOrder); err != nil {
		t.Fatal(err)
	}
	orderRepository := &fakeOrderRepository{assigned: []*order.Order{deliveryOrder}}
	courierRepository := &fakeCourierRepository{byID: map[uuid.UUID]*courier.Courier{aggregate.ID(): aggregate}}
	handler, err := NewMoveCouriersCommandHandler(fakeUnitOfWork{}, orderRepository, courierRepository)
	if err != nil {
		t.Fatal(err)
	}
	if err = handler.Handle(context.Background(), MoveCouriersCommand{}); err != nil {
		t.Fatal(err)
	}
	if deliveryOrder.Status() != order.StatusCompleted || aggregate.StoragePlaces()[0].OrderID() != nil {
		t.Errorf("delivery was not completed: order=%s, storage=%v", deliveryOrder.Status(), aggregate.StoragePlaces()[0].OrderID())
	}
}

type fakeUnitOfWork struct{}

func (fakeUnitOfWork) Do(_ context.Context, fn func(tx ports.Tx) error) error {
	return fn(&fakeTx{})
}

type fakeTx struct {
	events []ddd.DomainEvent
}

func (t *fakeTx) DB() *gorm.DB                   { return nil }
func (t *fakeTx) AddEvent(event ddd.DomainEvent) { t.events = append(t.events, event) }
func (t *fakeTx) Events() []ddd.DomainEvent      { return t.events }

type fakeOrderRepository struct {
	added    *order.Order
	updated  *order.Order
	created  *order.Order
	assigned []*order.Order
}

func (r *fakeOrderRepository) Add(_ context.Context, _ ports.Tx, aggregate *order.Order) error {
	r.added = aggregate
	return nil
}
func (r *fakeOrderRepository) Update(_ context.Context, _ ports.Tx, aggregate *order.Order) error {
	r.updated = aggregate
	return nil
}
func (r *fakeOrderRepository) Get(_ context.Context, _ ports.Tx, _ uuid.UUID) (*order.Order, error) {
	return nil, nil
}
func (r *fakeOrderRepository) GetAllCreated(_ context.Context, _ ports.Tx) ([]*order.Order, error) {
	if r.created != nil {
		return []*order.Order{r.created}, nil
	}
	return nil, nil
}
func (r *fakeOrderRepository) GetAllAssigned(_ context.Context, _ ports.Tx) ([]*order.Order, error) {
	return r.assigned, nil
}
func (r *fakeOrderRepository) GetAllNotCompleted(_ context.Context, _ ports.Tx) ([]*order.Order, error) {
	return nil, nil
}

type fakeCourierRepository struct {
	free    []*courier.Courier
	byID    map[uuid.UUID]*courier.Courier
	updated *courier.Courier
}

func (r *fakeCourierRepository) Add(_ context.Context, _ ports.Tx, _ *courier.Courier) error {
	return nil
}
func (r *fakeCourierRepository) Update(_ context.Context, _ ports.Tx, aggregate *courier.Courier) error {
	r.updated = aggregate
	return nil
}
func (r *fakeCourierRepository) Get(_ context.Context, _ ports.Tx, id uuid.UUID) (*courier.Courier, error) {
	return r.byID[id], nil
}
func (r *fakeCourierRepository) GetAll(_ context.Context, _ ports.Tx) ([]*courier.Courier, error) {
	return r.free, nil
}
func (r *fakeCourierRepository) GetAllFree(_ context.Context, _ ports.Tx) ([]*courier.Courier, error) {
	return r.free, nil
}
