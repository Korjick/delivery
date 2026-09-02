package order

import (
	"delivery/internal/core/domain/kernel"
	"delivery/internal/pkg/ddd"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestNewOrder(t *testing.T) {
	location := kernel.MustNewLocation(1, 1)
	type args struct {
		orderID  uuid.UUID
		location kernel.Location
		volume   int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "correct_order",
			args: args{orderID: uuid.NameSpaceDNS, location: location, volume: 1},
		},
		{
			name:    "incorrect_empty_order_id",
			args:    args{location: location, volume: 1},
			wantErr: true,
		},
		{
			name:    "incorrect_empty_location",
			args:    args{orderID: uuid.NameSpaceDNS, volume: 1},
			wantErr: true,
		},
		{
			name:    "incorrect_zero_volume",
			args:    args{orderID: uuid.NameSpaceDNS, location: location},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewOrder(tt.args.orderID, tt.args.location, tt.args.volume)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.ID() != tt.args.orderID || got.CourierID() != nil || !reflect.DeepEqual(got.Location(), tt.args.location) || got.Volume() != tt.args.volume || got.Status() != StatusCreated {
				t.Errorf("NewOrder() got = %#v, want created order", got)
			}
		})
	}
}

func TestOrder_Equals(t *testing.T) {
	type fields struct {
		base *ddd.BaseAggregate[uuid.UUID]
	}
	type args struct {
		other *Order
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "equal_same_id",
			fields: fields{base: ddd.NewBaseAggregate(uuid.NameSpaceDNS)},
			args:   args{other: &Order{BaseAggregate: ddd.NewBaseAggregate(uuid.NameSpaceDNS), volume: 10, status: StatusCompleted}},
			want:   true,
		},
		{
			name:   "not_equal_different_id",
			fields: fields{base: ddd.NewBaseAggregate(uuid.NameSpaceDNS)},
			args:   args{other: &Order{BaseAggregate: ddd.NewBaseAggregate(uuid.NameSpaceURL)}},
		},
		{
			name:   "not_equal_nil_order",
			fields: fields{base: ddd.NewBaseAggregate(uuid.NameSpaceDNS)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{BaseAggregate: tt.fields.base}
			if got := o.Equals(tt.args.other); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrder_Assign(t *testing.T) {
	type fields struct {
		courierID *uuid.UUID
		status    Status
	}
	type args struct {
		courierID uuid.UUID
	}
	assignedCourierID := uuid.NameSpaceURL
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantCourierID *uuid.UUID
		wantStatus    Status
		wantErr       bool
	}{
		{
			name:          "assign_created_order",
			fields:        fields{status: StatusCreated},
			args:          args{courierID: assignedCourierID},
			wantCourierID: &assignedCourierID,
			wantStatus:    StatusAssigned,
		},
		{
			name:       "incorrect_empty_courier_id",
			fields:     fields{status: StatusCreated},
			wantStatus: StatusCreated,
			wantErr:    true,
		},
		{
			name:          "incorrect_already_assigned_order",
			fields:        fields{courierID: &assignedCourierID, status: StatusAssigned},
			args:          args{courierID: uuid.NameSpaceOID},
			wantCourierID: &assignedCourierID,
			wantStatus:    StatusAssigned,
			wantErr:       true,
		},
		{
			name:          "incorrect_completed_order",
			fields:        fields{courierID: &assignedCourierID, status: StatusCompleted},
			args:          args{courierID: uuid.NameSpaceOID},
			wantCourierID: &assignedCourierID,
			wantStatus:    StatusCompleted,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{BaseAggregate: ddd.NewBaseAggregate(uuid.New()), courierID: tt.fields.courierID, status: tt.fields.status}
			if err := o.Assign(tt.args.courierID); (err != nil) != tt.wantErr {
				t.Errorf("Assign() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := o.CourierID(); !equalUUIDPointers(got, tt.wantCourierID) {
				t.Errorf("CourierID() = %v, want %v", got, tt.wantCourierID)
			}
			if got := o.Status(); got != tt.wantStatus {
				t.Errorf("Status() = %v, want %v", got, tt.wantStatus)
			}
		})
	}
}

func TestOrder_Complete(t *testing.T) {
	type fields struct {
		courierID *uuid.UUID
		status    Status
	}
	assignedCourierID := uuid.NameSpaceDNS
	tests := []struct {
		name       string
		fields     fields
		wantStatus Status
		wantErr    bool
	}{
		{
			name:       "complete_assigned_order",
			fields:     fields{courierID: &assignedCourierID, status: StatusAssigned},
			wantStatus: StatusCompleted,
		},
		{
			name:       "incorrect_created_order",
			fields:     fields{status: StatusCreated},
			wantStatus: StatusCreated,
			wantErr:    true,
		},
		{
			name:       "incorrect_completed_order",
			fields:     fields{courierID: &assignedCourierID, status: StatusCompleted},
			wantStatus: StatusCompleted,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{BaseAggregate: ddd.NewBaseAggregate(uuid.New()), courierID: tt.fields.courierID, status: tt.fields.status}
			if err := o.Complete(); (err != nil) != tt.wantErr {
				t.Errorf("Complete() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := o.Status(); got != tt.wantStatus {
				t.Errorf("Status() = %v, want %v", got, tt.wantStatus)
			}
		})
	}
}

func equalUUIDPointers(first, second *uuid.UUID) bool {
	if first == nil || second == nil {
		return first == second
	}
	return *first == *second
}
