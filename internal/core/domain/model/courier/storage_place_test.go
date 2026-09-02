package courier

import (
	"delivery/internal/pkg/ddd"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestNewStoragePlace(t *testing.T) {
	type args struct {
		name        string
		totalVolume int
	}
	tests := []struct {
		name            string
		args            args
		wantName        string
		wantTotalVolume int
		wantErr         bool
	}{
		{
			name:            "correct_storage_place",
			wantName:        "Backpack",
			wantTotalVolume: 10,
			args: args{
				name:        "Backpack",
				totalVolume: 10,
			},
		},
		{
			name:    "incorrect_empty_name",
			wantErr: true,
			args: args{
				totalVolume: 1,
			},
		},
		{
			name:    "incorrect_blank_name",
			wantErr: true,
			args: args{
				name:        "  ",
				totalVolume: 1,
			},
		},
		{
			name:    "incorrect_zero_total_volume",
			wantErr: true,
			args: args{
				name: "Backpack",
			},
		},
		{
			name:    "incorrect_negative_total_volume",
			wantErr: true,
			args: args{
				name:        "Backpack",
				totalVolume: -1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewStoragePlace(tt.args.name, tt.args.totalVolume)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewStoragePlace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.ID() == uuid.Nil {
				t.Error("NewStoragePlace() created an empty ID")
			}
			if got.Name() != tt.wantName || got.TotalVolume() != tt.wantTotalVolume || got.OrderID() != nil {
				t.Errorf("NewStoragePlace() got = %#v, want name %q, total volume %d and empty order ID", got, tt.wantName, tt.wantTotalVolume)
			}
		})
	}
}

func TestStoragePlace_Equals(t *testing.T) {
	type fields struct {
		base        *ddd.BaseEntity[uuid.UUID]
		name        string
		totalVolume int
		orderID     *uuid.UUID
	}
	type args struct {
		other *StoragePlace
	}
	sameID := uuid.NameSpaceDNS
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "equal_same_id_different_state",
			fields: fields{
				base:        ddd.NewBaseEntity(sameID),
				name:        "Backpack",
				totalVolume: 1,
			},
			args: args{
				other: &StoragePlace{
					base:        ddd.NewBaseEntity(sameID),
					name:        "Trunk",
					totalVolume: 100,
				},
			},
			want: true,
		},
		{
			name: "not_equal_different_id",
			fields: fields{
				base: ddd.NewBaseEntity(uuid.NameSpaceDNS),
			},
			args: args{
				other: &StoragePlace{base: ddd.NewBaseEntity(uuid.NameSpaceURL)},
			},
		},
		{
			name: "not_equal_nil_entity",
			fields: fields{
				base: ddd.NewBaseEntity(uuid.NameSpaceDNS),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StoragePlace{
				base:        tt.fields.base,
				name:        tt.fields.name,
				totalVolume: tt.fields.totalVolume,
				orderID:     tt.fields.orderID,
			}
			if got := s.Equals(tt.args.other); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStoragePlace_CanStore(t *testing.T) {
	type fields struct {
		base        *ddd.BaseEntity[uuid.UUID]
		name        string
		totalVolume int
		orderID     *uuid.UUID
	}
	type args struct {
		volume int
	}
	storedOrderID := uuid.NameSpaceDNS
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "can_store_order_with_equal_volume",
			fields: fields{
				base:        ddd.NewBaseEntity(uuid.New()),
				name:        "Backpack",
				totalVolume: 10,
			},
			args: args{volume: 10},
			want: true,
		},
		{
			name: "cannot_store_order_with_too_large_volume",
			fields: fields{
				base:        ddd.NewBaseEntity(uuid.New()),
				name:        "Backpack",
				totalVolume: 10,
			},
			args: args{volume: 11},
		},
		{
			name: "can_check_volume_in_occupied_storage_place",
			fields: fields{
				base:        ddd.NewBaseEntity(uuid.New()),
				name:        "Backpack",
				totalVolume: 10,
				orderID:     &storedOrderID,
			},
			args: args{volume: 1},
			want: true,
		},
		{
			name: "incorrect_zero_volume",
			fields: fields{
				base:        ddd.NewBaseEntity(uuid.New()),
				totalVolume: 10,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StoragePlace{
				base:        tt.fields.base,
				name:        tt.fields.name,
				totalVolume: tt.fields.totalVolume,
				orderID:     tt.fields.orderID,
			}
			got, err := s.CanStore(tt.args.volume)
			if (err != nil) != tt.wantErr {
				t.Errorf("CanStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CanStore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStoragePlace_Store(t *testing.T) {
	type fields struct {
		base        *ddd.BaseEntity[uuid.UUID]
		name        string
		totalVolume int
		orderID     *uuid.UUID
	}
	type args struct {
		orderID uuid.UUID
		volume  int
	}
	storedOrderID := uuid.NameSpaceDNS
	newOrderID := uuid.NameSpaceURL
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantOrderID *uuid.UUID
		wantErr     bool
	}{
		{
			name: "store_order",
			fields: fields{
				base:        ddd.NewBaseEntity(uuid.New()),
				name:        "Backpack",
				totalVolume: 10,
			},
			args:        args{orderID: newOrderID, volume: 10},
			wantOrderID: &newOrderID,
		},
		{
			name: "incorrect_empty_order_id",
			fields: fields{
				base:        ddd.NewBaseEntity(uuid.New()),
				totalVolume: 10,
			},
			args:    args{volume: 1},
			wantErr: true,
		},
		{
			name: "incorrect_zero_volume",
			fields: fields{
				base:        ddd.NewBaseEntity(uuid.New()),
				totalVolume: 10,
			},
			args:    args{orderID: newOrderID},
			wantErr: true,
		},
		{
			name: "incorrect_too_large_volume",
			fields: fields{
				base:        ddd.NewBaseEntity(uuid.New()),
				totalVolume: 10,
			},
			args:    args{orderID: newOrderID, volume: 11},
			wantErr: true,
		},
		{
			name: "incorrect_occupied_storage_place",
			fields: fields{
				base:        ddd.NewBaseEntity(uuid.New()),
				totalVolume: 10,
				orderID:     &storedOrderID,
			},
			args:        args{orderID: newOrderID, volume: 1},
			wantOrderID: &storedOrderID,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StoragePlace{
				base:        tt.fields.base,
				name:        tt.fields.name,
				totalVolume: tt.fields.totalVolume,
				orderID:     tt.fields.orderID,
			}
			if err := s.Store(tt.args.orderID, tt.args.volume); (err != nil) != tt.wantErr {
				t.Errorf("Store() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := s.OrderID(); !reflect.DeepEqual(got, tt.wantOrderID) {
				t.Errorf("OrderID() = %v, want %v", got, tt.wantOrderID)
			}
		})
	}
}

func TestStoragePlace_Clear(t *testing.T) {
	type fields struct {
		base        *ddd.BaseEntity[uuid.UUID]
		name        string
		totalVolume int
		orderID     *uuid.UUID
	}
	type args struct {
		orderID uuid.UUID
	}
	storedOrderID := uuid.NameSpaceDNS
	otherOrderID := uuid.NameSpaceURL
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantOrderID *uuid.UUID
		wantErr     bool
	}{
		{
			name: "clear_stored_order",
			fields: fields{
				base:        ddd.NewBaseEntity(uuid.New()),
				totalVolume: 10,
				orderID:     &storedOrderID,
			},
			args: args{orderID: storedOrderID},
		},
		{
			name: "incorrect_empty_order_id",
			fields: fields{
				base:        ddd.NewBaseEntity(uuid.New()),
				totalVolume: 10,
				orderID:     &storedOrderID,
			},
			wantOrderID: &storedOrderID,
			wantErr:     true,
		},
		{
			name: "incorrect_not_stored_order_id",
			fields: fields{
				base:        ddd.NewBaseEntity(uuid.New()),
				totalVolume: 10,
				orderID:     &storedOrderID,
			},
			args:        args{orderID: otherOrderID},
			wantOrderID: &storedOrderID,
			wantErr:     true,
		},
		{
			name: "incorrect_empty_storage_place",
			fields: fields{
				base:        ddd.NewBaseEntity(uuid.New()),
				totalVolume: 10,
			},
			args:    args{orderID: storedOrderID},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StoragePlace{
				base:        tt.fields.base,
				name:        tt.fields.name,
				totalVolume: tt.fields.totalVolume,
				orderID:     tt.fields.orderID,
			}
			if err := s.Clear(tt.args.orderID); (err != nil) != tt.wantErr {
				t.Errorf("Clear() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := s.OrderID(); !reflect.DeepEqual(got, tt.wantOrderID) {
				t.Errorf("OrderID() = %v, want %v", got, tt.wantOrderID)
			}
		})
	}
}
