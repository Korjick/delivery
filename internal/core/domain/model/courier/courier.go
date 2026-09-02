package courier

import (
	"delivery/internal/core/domain/kernel"
	"delivery/internal/core/domain/model/order"
	"delivery/internal/pkg/ddd"
	"delivery/internal/pkg/errs"
	"delivery/internal/pkg/mathx"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const (
	defaultStoragePlaceName   = "Сумка"
	defaultStoragePlaceVolume = 10
)

var (
	ErrOrderCannotBeTaken          = errors.New("courier cannot take order")
	ErrOrderIsNotStoredByCourier   = errors.New("order is not stored by courier")
	ErrOrderAssignedToOtherCourier = errors.New("order is assigned to another courier")
)

type Courier struct {
	*ddd.BaseAggregate[uuid.UUID]
	name          string
	speed         int
	location      kernel.Location
	storagePlaces []*StoragePlace
}

func NewCourier(name string, speed int, location kernel.Location) (*Courier, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errs.NewValueIsRequired("name")
	}
	if speed <= 0 {
		return nil, errs.NewValueMustBeGreaterOrEqual("speed", speed, 1)
	}
	if location.IsEmpty() {
		return nil, errs.NewValueIsRequired("location")
	}

	storagePlace, err := NewStoragePlace(defaultStoragePlaceName, defaultStoragePlaceVolume)
	if err != nil {
		return nil, fmt.Errorf("create default storage place: %w", err)
	}

	return &Courier{
		BaseAggregate: ddd.NewBaseAggregate(uuid.New()),
		name:          name,
		speed:         speed,
		location:      location,
		storagePlaces: []*StoragePlace{storagePlace},
	}, nil
}

func RestoreCourier(id uuid.UUID, name string, speed int,
	location kernel.Location, storagePlaces []*StoragePlace) *Courier {
	return &Courier{
		BaseAggregate: ddd.NewBaseAggregate(id),
		name:          name,
		speed:         speed,
		location:      location,
		storagePlaces: append([]*StoragePlace(nil), storagePlaces...),
	}
}

func (c *Courier) Equals(other *Courier) bool {
	return other != nil && c.Equal(other.BaseAggregate)
}

func (c *Courier) Name() string {
	return c.name
}

func (c *Courier) Speed() int {
	return c.speed
}

func (c *Courier) Location() kernel.Location {
	return c.location
}

func (c *Courier) StoragePlaces() []*StoragePlace {
	return append([]*StoragePlace(nil), c.storagePlaces...)
}

func (c *Courier) AddStoragePlace(name string, volume int) error {
	storagePlace, err := NewStoragePlace(name, volume)
	if err != nil {
		return err
	}

	c.storagePlaces = append(c.storagePlaces, storagePlace)
	return nil
}

func (c *Courier) CanTakeOrder(deliveryOrder *order.Order) (bool, error) {
	if deliveryOrder == nil {
		return false, errs.NewValueIsRequired("order")
	}

	for _, storagePlace := range c.storagePlaces {
		if storagePlace.isOccupied() {
			continue
		}
		canStore, err := storagePlace.CanStore(deliveryOrder.Volume())
		if err != nil {
			return false, err
		}
		if canStore {
			return true, nil
		}
	}

	return false, nil
}

func (c *Courier) TakeOrder(deliveryOrder *order.Order) error {
	canTake, err := c.CanTakeOrder(deliveryOrder)
	if err != nil {
		return err
	}
	if !canTake {
		return ErrOrderCannotBeTaken
	}

	if err = deliveryOrder.Assign(c.ID()); err != nil {
		return err
	}
	storagePlace, err := c.firstAvailableStoragePlace(deliveryOrder.Volume())
	if err != nil {
		return err
	}
	return storagePlace.Store(deliveryOrder.ID(), deliveryOrder.Volume())
}

func (c *Courier) CompleteOrder(deliveryOrder *order.Order) error {
	if deliveryOrder == nil {
		return errs.NewValueIsRequired("order")
	}
	courierID := deliveryOrder.CourierID()
	if courierID == nil || *courierID != c.ID() {
		return ErrOrderAssignedToOtherCourier
	}

	storagePlace, err := c.findStoragePlaceByOrderID(deliveryOrder.ID())
	if err != nil {
		return err
	}
	if err = deliveryOrder.Complete(); err != nil {
		return err
	}
	return storagePlace.Clear(deliveryOrder.ID())
}

func (c *Courier) CalculateTimeToOrder(target kernel.Location) (float64, error) {
	if target.IsEmpty() {
		return 0, errs.NewValueIsRequired("target")
	}

	distance := c.location.Distance(&target)
	return float64(distance) / float64(c.speed), nil
}

func (c *Courier) Move(target kernel.Location) error {
	if target.IsEmpty() {
		return errs.NewValueIsRequired("target")
	}

	dx := target.X() - c.location.X()
	dy := target.Y() - c.location.Y()
	remainingRange := c.speed

	stepX := limitedStep(dx, remainingRange)
	remainingRange -= mathx.Abs(stepX)
	stepY := limitedStep(dy, remainingRange)

	newLocation, err := kernel.NewLocation(c.location.X()+stepX, c.location.Y()+stepY)
	if err != nil {
		return err
	}
	c.location = *newLocation
	return nil
}

func (c *Courier) findStoragePlaceByOrderID(orderID uuid.UUID) (*StoragePlace, error) {
	if orderID == uuid.Nil {
		return nil, errs.NewValueIsRequired("orderID")
	}

	for _, storagePlace := range c.storagePlaces {
		storedOrderID := storagePlace.OrderID()
		if storedOrderID != nil && *storedOrderID == orderID {
			return storagePlace, nil
		}
	}
	return nil, ErrOrderIsNotStoredByCourier
}

func (c *Courier) firstAvailableStoragePlace(volume int) (*StoragePlace, error) {
	for _, storagePlace := range c.storagePlaces {
		if storagePlace.isOccupied() {
			continue
		}
		canStore, err := storagePlace.CanStore(volume)
		if err != nil {
			return nil, err
		}
		if canStore {
			return storagePlace, nil
		}
	}
	return nil, ErrOrderCannotBeTaken
}

func limitedStep(delta, limit int) int {
	if delta > limit {
		return limit
	}
	if delta < -limit {
		return -limit
	}
	return delta
}
