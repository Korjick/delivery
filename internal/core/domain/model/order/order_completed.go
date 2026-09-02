package order

import (
	"delivery/internal/pkg/ddd"
	"reflect"

	"github.com/google/uuid"
)

var _ ddd.DomainEvent = (*CompletedDomainEvent)(nil)

type CompletedDomainEvent struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	OrderID uuid.UUID `json:"orderId"`
}

func (e *CompletedDomainEvent) GetID() uuid.UUID { return e.ID }

func (e *CompletedDomainEvent) GetName() string { return e.Name }

func NewCompletedDomainEvent(orderID uuid.UUID) *CompletedDomainEvent {
	event := &CompletedDomainEvent{ID: uuid.New(), OrderID: orderID}
	event.Name = reflect.TypeOf(*event).Name()
	return event
}

func NewEmptyCompletedDomainEvent() ddd.DomainEvent {
	event := &CompletedDomainEvent{}
	event.Name = reflect.TypeOf(*event).Name()
	return event
}
