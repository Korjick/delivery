package order

import (
	"delivery/internal/core/domain/kernel"
	"delivery/internal/pkg/ddd"
	"delivery/internal/pkg/errs"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrOrderAlreadyAssigned  = errors.New("order is already assigned")
	ErrOrderNotAssigned      = errors.New("order is not assigned")
	ErrOrderAlreadyCompleted = errors.New("order is already completed")
)

type Order struct {
	*ddd.BaseAggregate[uuid.UUID]
	courierID *uuid.UUID
	location  kernel.Location
	volume    int
	status    Status
}

func (o *Order) Equals(other *Order) bool {
	return other != nil && o.Equal(other.BaseAggregate)
}

func (o *Order) CourierID() *uuid.UUID {
	if o.courierID == nil {
		return nil
	}

	courierID := *o.courierID
	return &courierID
}

func (o *Order) Location() kernel.Location {
	return o.location
}

func (o *Order) Volume() int {
	return o.volume
}

func (o *Order) Status() Status {
	return o.status
}

func (o *Order) Assign(courierID uuid.UUID) error {
	if courierID == uuid.Nil {
		return errs.NewValueIsRequired("courierID")
	}
	if o.status == StatusCompleted {
		return ErrOrderAlreadyCompleted
	}
	if o.status == StatusAssigned {
		return ErrOrderAlreadyAssigned
	}

	assignedCourierID := courierID
	o.courierID = &assignedCourierID
	o.status = StatusAssigned
	return nil
}

func (o *Order) Complete() error {
	if o.status == StatusCompleted {
		return ErrOrderAlreadyCompleted
	}
	if o.status != StatusAssigned || o.courierID == nil {
		return ErrOrderNotAssigned
	}

	o.status = StatusCompleted
	o.RaiseDomainEvent(NewCompletedDomainEvent(o.ID()))
	return nil
}

func NewOrder(orderID uuid.UUID, location kernel.Location, volume int) (*Order, error) {
	if orderID == uuid.Nil {
		return nil, errs.NewValueIsRequired("orderID")
	}
	if location.IsEmpty() {
		return nil, errs.NewValueIsRequired("location")
	}
	if volume <= 0 {
		return nil, errs.NewValueIsRequired("volume")
	}

	return &Order{
		BaseAggregate: ddd.NewBaseAggregate(orderID),
		location:      location,
		volume:        volume,
		status:        StatusCreated,
	}, nil
}

func RestoreOrder(orderID uuid.UUID, courierID *uuid.UUID, location kernel.Location, volume int, status Status) *Order {
	var restoredCourierID *uuid.UUID
	if courierID != nil {
		courierIDCopy := *courierID
		restoredCourierID = &courierIDCopy
	}

	return &Order{
		BaseAggregate: ddd.NewBaseAggregate(orderID),
		courierID:     restoredCourierID,
		location:      location,
		volume:        volume,
		status:        status,
	}
}
