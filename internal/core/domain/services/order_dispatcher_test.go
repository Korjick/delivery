package services

import (
	"delivery/internal/core/domain/kernel"
	"delivery/internal/core/domain/model/courier"
	"delivery/internal/core/domain/model/order"
	"delivery/internal/pkg/errs"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestOrderDispatcher_Dispatch(t *testing.T) {
	type courierSpec struct {
		name     string
		speed    int
		location kernel.Location
		occupied bool
	}
	tests := []struct {
		name               string
		orderVolume        int
		orderAssigned      bool
		couriers           []courierSpec
		wantWinnerPosition int
		wantErr            bool
	}{
		{
			name:        "assign_order_to_fastest_courier",
			orderVolume: 1,
			couriers: []courierSpec{
				{name: "Slow", speed: 1, location: kernel.MustNewLocation(1, 1)},
				{name: "Fast", speed: 2, location: kernel.MustNewLocation(1, 1)},
			},
			wantWinnerPosition: 1,
		},
		{
			name:        "skip_faster_courier_with_occupied_storage",
			orderVolume: 1,
			couriers: []courierSpec{
				{name: "Busy", speed: 10, location: kernel.MustNewLocation(1, 1), occupied: true},
				{name: "Available", speed: 1, location: kernel.MustNewLocation(1, 1)},
			},
			wantWinnerPosition: 1,
		},
		{
			name:        "no_suitable_courier_for_large_order",
			orderVolume: 11,
			couriers: []courierSpec{
				{name: "Courier", speed: 1, location: kernel.MustNewLocation(1, 1)},
			},
			wantWinnerPosition: -1,
			wantErr:            true,
		},
		{
			name:               "incorrect_assigned_order",
			orderVolume:        1,
			orderAssigned:      true,
			wantWinnerPosition: -1,
			wantErr:            true,
		},
		{
			name:               "incorrect_empty_couriers",
			orderVolume:        1,
			wantWinnerPosition: -1,
			wantErr:            true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deliveryOrder := newDispatchOrder(t, tt.orderVolume)
			if tt.orderAssigned {
				if err := deliveryOrder.Assign(uuid.NameSpaceDNS); err != nil {
					t.Fatal(err)
				}
			}

			couriers := make([]*courier.Courier, 0, len(tt.couriers))
			for _, spec := range tt.couriers {
				candidate, err := courier.NewCourier(spec.name, spec.speed, spec.location)
				if err != nil {
					t.Fatal(err)
				}
				if spec.occupied {
					if err = candidate.StoragePlaces()[0].Store(uuid.New(), 1); err != nil {
						t.Fatal(err)
					}
				}
				couriers = append(couriers, candidate)
			}

			got, err := NewOrderDispatcher().Dispatch(deliveryOrder, couriers)
			if (err != nil) != tt.wantErr {
				t.Errorf("Dispatch() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if got != nil {
					t.Errorf("Dispatch() courier = %v, want nil", got)
				}
				return
			}
			if got != couriers[tt.wantWinnerPosition] {
				t.Errorf("Dispatch() courier = %v, want %v", got, couriers[tt.wantWinnerPosition])
			}
			if deliveryOrder.Status() != order.StatusAssigned || deliveryOrder.CourierID() == nil || *deliveryOrder.CourierID() != got.ID() {
				t.Errorf("Dispatch() did not assign order to winner: %#v", deliveryOrder)
			}
			storedOrderID := got.StoragePlaces()[0].OrderID()
			if storedOrderID == nil || *storedOrderID != deliveryOrder.ID() {
				t.Errorf("Dispatch() did not store order: %v", storedOrderID)
			}
		})
	}
}

func TestOrderDispatcher_DispatchNilOrder(t *testing.T) {
	got, err := NewOrderDispatcher().Dispatch(nil, nil)
	if got != nil {
		t.Errorf("Dispatch() courier = %v, want nil", got)
	}
	if !errors.Is(err, errs.ErrValueIsRequired) {
		t.Errorf("Dispatch() error = %v, want required value error", err)
	}
}

func newDispatchOrder(t *testing.T, volume int) *order.Order {
	t.Helper()
	deliveryOrder, err := order.NewOrder(uuid.New(), kernel.MustNewLocation(5, 5), volume)
	if err != nil {
		t.Fatal(err)
	}
	return deliveryOrder
}
