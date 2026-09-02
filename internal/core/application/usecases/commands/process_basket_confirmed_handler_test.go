package commands

import (
	"context"
	"delivery/internal/core/domain/kernel"
	"delivery/internal/core/domain/model/order"
	"delivery/internal/core/ports"
	"testing"

	"github.com/google/uuid"
)

func TestProcessBasketConfirmedCommandHandler_Handle_IsIdempotent(t *testing.T) {
	inboxRepository := &recordingInboxRepository{processed: make(map[string]struct{})}
	orderRepository := &recordingOrderRepository{}
	handler, err := NewProcessBasketConfirmedCommandHandler(
		fakeUnitOfWork{},
		inboxRepository,
		orderRepository,
		fakeGeoClient{location: kernel.MustNewLocation(3, 4)},
	)
	if err != nil {
		t.Fatal(err)
	}

	command := ProcessBasketConfirmedCommand{
		MessageID: "basket.confirmed:0:42",
		OrderID:   uuid.New(),
		Street:    "Lenina",
		Volume:    3,
	}
	if err = handler.Handle(context.Background(), command); err != nil {
		t.Fatal(err)
	}
	if err = handler.Handle(context.Background(), command); err != nil {
		t.Fatal(err)
	}

	if orderRepository.addCalls != 1 {
		t.Errorf("order Add calls = %d, want 1", orderRepository.addCalls)
	}
	if orderRepository.added == nil || orderRepository.added.ID() != command.OrderID {
		t.Errorf("created order = %#v, want %s", orderRepository.added, command.OrderID)
	}
}

func TestProcessBasketConfirmedCommandHandler_Handle_RequiresMessageID(t *testing.T) {
	handler, err := NewProcessBasketConfirmedCommandHandler(
		fakeUnitOfWork{},
		&recordingInboxRepository{processed: make(map[string]struct{})},
		&recordingOrderRepository{},
		fakeGeoClient{location: kernel.MustNewLocation(3, 4)},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = handler.Handle(context.Background(), ProcessBasketConfirmedCommand{
		OrderID: uuid.New(),
		Street:  "Lenina",
		Volume:  3,
	})
	if err == nil {
		t.Fatal("Handle() error = nil, want missing messageID error")
	}
}

type recordingInboxRepository struct {
	processed map[string]struct{}
}

func (r *recordingInboxRepository) TryAdd(_ context.Context, _ ports.Tx, messageID string) (bool, error) {
	if _, exists := r.processed[messageID]; exists {
		return false, nil
	}
	r.processed[messageID] = struct{}{}
	return true, nil
}

type recordingOrderRepository struct {
	added    *order.Order
	addCalls int
}

func (r *recordingOrderRepository) Add(_ context.Context, _ ports.Tx, aggregate *order.Order) error {
	r.addCalls++
	r.added = aggregate
	return nil
}

func (*recordingOrderRepository) Update(_ context.Context, _ ports.Tx, _ *order.Order) error {
	return nil
}

func (*recordingOrderRepository) Get(_ context.Context, _ ports.Tx, _ uuid.UUID) (*order.Order, error) {
	return nil, nil
}

func (*recordingOrderRepository) GetAllCreated(_ context.Context, _ ports.Tx) ([]*order.Order, error) {
	return nil, nil
}

func (*recordingOrderRepository) GetAllAssigned(_ context.Context, _ ports.Tx) ([]*order.Order, error) {
	return nil, nil
}

func (*recordingOrderRepository) GetAllNotCompleted(_ context.Context, _ ports.Tx) ([]*order.Order, error) {
	return nil, nil
}
