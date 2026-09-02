package courier

import (
	"delivery/internal/core/domain/kernel"
	"delivery/internal/core/domain/model/order"
	"delivery/internal/pkg/ddd"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestNewCourier(t *testing.T) {
	location := kernel.MustNewLocation(1, 1)
	type args struct {
		name     string
		speed    int
		location kernel.Location
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "correct_courier",
			args: args{name: "Alex", speed: 2, location: location},
		},
		{
			name:    "incorrect_empty_name",
			args:    args{speed: 1, location: location},
			wantErr: true,
		},
		{
			name:    "incorrect_zero_speed",
			args:    args{name: "Alex", location: location},
			wantErr: true,
		},
		{
			name:    "incorrect_empty_location",
			args:    args{name: "Alex", speed: 1},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCourier(tt.args.name, tt.args.speed, tt.args.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCourier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.ID() == uuid.Nil || got.Name() != tt.args.name || got.Speed() != tt.args.speed || !reflect.DeepEqual(got.Location(), tt.args.location) {
				t.Errorf("NewCourier() got = %#v, want valid courier", got)
			}
			storagePlaces := got.StoragePlaces()
			if len(storagePlaces) != 1 || storagePlaces[0].Name() != defaultStoragePlaceName || storagePlaces[0].TotalVolume() != defaultStoragePlaceVolume {
				t.Errorf("NewCourier() storage places = %#v, want default bag", storagePlaces)
			}
		})
	}
}

func TestCourier_Equals(t *testing.T) {
	type fields struct {
		base *ddd.BaseAggregate[uuid.UUID]
	}
	type args struct {
		other *Courier
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
			args:   args{other: &Courier{BaseAggregate: ddd.NewBaseAggregate(uuid.NameSpaceDNS), name: "Other"}},
			want:   true,
		},
		{
			name:   "not_equal_different_id",
			fields: fields{base: ddd.NewBaseAggregate(uuid.NameSpaceDNS)},
			args:   args{other: &Courier{BaseAggregate: ddd.NewBaseAggregate(uuid.NameSpaceURL)}},
		},
		{
			name:   "not_equal_nil_courier",
			fields: fields{base: ddd.NewBaseAggregate(uuid.NameSpaceDNS)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Courier{BaseAggregate: tt.fields.base}
			if got := c.Equals(tt.args.other); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCourier_AddStoragePlace(t *testing.T) {
	type args struct {
		name   string
		volume int
	}
	tests := []struct {
		name      string
		args      args
		wantCount int
		wantErr   bool
	}{
		{
			name:      "add_storage_place",
			args:      args{name: "Trunk", volume: 100},
			wantCount: 2,
		},
		{
			name:      "incorrect_empty_name",
			args:      args{volume: 1},
			wantCount: 1,
			wantErr:   true,
		},
		{
			name:      "incorrect_zero_volume",
			args:      args{name: "Trunk"},
			wantCount: 1,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestCourier(t, 1)
			if err := c.AddStoragePlace(tt.args.name, tt.args.volume); (err != nil) != tt.wantErr {
				t.Errorf("AddStoragePlace() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := len(c.StoragePlaces()); got != tt.wantCount {
				t.Errorf("StoragePlaces() count = %v, want %v", got, tt.wantCount)
			}
		})
	}
}

func TestCourier_CanTakeOrder(t *testing.T) {
	tests := []struct {
		name        string
		orderVolume int
		occupied    bool
		nilOrder    bool
		want        bool
		wantErr     bool
	}{
		{
			name:        "can_take_order",
			orderVolume: 10,
			want:        true,
		},
		{
			name:        "cannot_take_too_large_order",
			orderVolume: 11,
		},
		{
			name:        "cannot_take_order_when_storage_is_occupied",
			orderVolume: 1,
			occupied:    true,
		},
		{
			name:     "incorrect_nil_order",
			nilOrder: true,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestCourier(t, 1)
			if tt.occupied {
				if err := c.StoragePlaces()[0].Store(uuid.NameSpaceDNS, 1); err != nil {
					t.Fatal(err)
				}
			}
			var deliveryOrder *order.Order
			if !tt.nilOrder {
				deliveryOrder = newTestOrder(t, tt.orderVolume)
			}
			got, err := c.CanTakeOrder(deliveryOrder)
			if (err != nil) != tt.wantErr {
				t.Errorf("CanTakeOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("CanTakeOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCourier_TakeOrder(t *testing.T) {
	tests := []struct {
		name            string
		orderVolume     int
		occupied        bool
		alreadyAssigned bool
		wantStatus      order.Status
		wantStored      bool
		wantErr         bool
	}{
		{
			name:        "take_order",
			orderVolume: 10,
			wantStatus:  order.StatusAssigned,
			wantStored:  true,
		},
		{
			name:        "cannot_take_too_large_order",
			orderVolume: 11,
			wantStatus:  order.StatusCreated,
			wantErr:     true,
		},
		{
			name:        "cannot_take_order_in_occupied_storage_place",
			orderVolume: 1,
			occupied:    true,
			wantStatus:  order.StatusCreated,
			wantErr:     true,
		},
		{
			name:            "cannot_take_already_assigned_order",
			orderVolume:     1,
			alreadyAssigned: true,
			wantStatus:      order.StatusAssigned,
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestCourier(t, 1)
			deliveryOrder := newTestOrder(t, tt.orderVolume)
			if tt.occupied {
				if err := c.StoragePlaces()[0].Store(uuid.NameSpaceDNS, 1); err != nil {
					t.Fatal(err)
				}
			}
			if tt.alreadyAssigned {
				if err := deliveryOrder.Assign(uuid.NameSpaceDNS); err != nil {
					t.Fatal(err)
				}
			}
			if err := c.TakeOrder(deliveryOrder); (err != nil) != tt.wantErr {
				t.Errorf("TakeOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := deliveryOrder.Status(); got != tt.wantStatus {
				t.Errorf("Status() = %v, want %v", got, tt.wantStatus)
			}
			stored := c.StoragePlaces()[0].OrderID() != nil && *c.StoragePlaces()[0].OrderID() == deliveryOrder.ID()
			if stored != tt.wantStored {
				t.Errorf("stored order = %v, want %v", stored, tt.wantStored)
			}
		})
	}
}

func TestCourier_CompleteOrder(t *testing.T) {
	tests := []struct {
		name            string
		orderTaken      bool
		assignedToOther bool
		wantStatus      order.Status
		wantStored      bool
		wantErr         bool
	}{
		{
			name:       "complete_taken_order",
			orderTaken: true,
			wantStatus: order.StatusCompleted,
		},
		{
			name:       "incorrect_order_not_stored_by_courier",
			wantStatus: order.StatusAssigned,
			wantErr:    true,
		},
		{
			name:            "incorrect_order_assigned_to_other_courier",
			assignedToOther: true,
			wantStatus:      order.StatusAssigned,
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestCourier(t, 1)
			deliveryOrder := newTestOrder(t, 1)
			if tt.orderTaken {
				if err := c.TakeOrder(deliveryOrder); err != nil {
					t.Fatal(err)
				}
			}
			if !tt.orderTaken {
				assignedCourierID := c.ID()
				if tt.assignedToOther {
					assignedCourierID = uuid.NameSpaceDNS
				}
				if err := deliveryOrder.Assign(assignedCourierID); err != nil {
					t.Fatal(err)
				}
			}
			if err := c.CompleteOrder(deliveryOrder); (err != nil) != tt.wantErr {
				t.Errorf("CompleteOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := deliveryOrder.Status(); got != tt.wantStatus {
				t.Errorf("Status() = %v, want %v", got, tt.wantStatus)
			}
			stored := c.StoragePlaces()[0].OrderID() != nil
			if stored != tt.wantStored {
				t.Errorf("stored order = %v, want %v", stored, tt.wantStored)
			}
		})
	}
}

func TestCourier_CalculateTimeToOrder(t *testing.T) {
	type args struct {
		target kernel.Location
	}
	tests := []struct {
		name    string
		speed   int
		args    args
		want    float64
		wantErr bool
	}{
		{
			name:  "calculate_time",
			speed: 2,
			args:  args{target: kernel.MustNewLocation(5, 5)},
			want:  4,
		},
		{
			name:    "incorrect_empty_target",
			speed:   1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestCourier(t, tt.speed)
			got, err := c.CalculateTimeToOrder(tt.args.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateTimeToOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("CalculateTimeToOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCourier_Move(t *testing.T) {
	type args struct {
		target kernel.Location
	}
	tests := []struct {
		name    string
		speed   int
		args    args
		want    kernel.Location
		wantErr bool
	}{
		{
			name:  "move_by_speed_in_horizontal_direction_first",
			speed: 2,
			args:  args{target: kernel.MustNewLocation(5, 5)},
			want:  kernel.MustNewLocation(3, 1),
		},
		{
			name:  "move_to_near_location",
			speed: 3,
			args:  args{target: kernel.MustNewLocation(2, 2)},
			want:  kernel.MustNewLocation(2, 2),
		},
		{
			name:    "incorrect_empty_target",
			speed:   1,
			want:    kernel.MustNewLocation(1, 1),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestCourier(t, tt.speed)
			if err := c.Move(tt.args.target); (err != nil) != tt.wantErr {
				t.Errorf("Move() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := c.Location(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Location() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newTestCourier(t *testing.T, speed int) *Courier {
	t.Helper()
	c, err := NewCourier("Alex", speed, kernel.MustNewLocation(1, 1))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func newTestOrder(t *testing.T, volume int) *order.Order {
	t.Helper()
	deliveryOrder, err := order.NewOrder(uuid.New(), kernel.MustNewLocation(5, 5), volume)
	if err != nil {
		t.Fatal(err)
	}
	return deliveryOrder
}
