package courier

import (
	"delivery/internal/pkg/ddd"
	"delivery/internal/pkg/errs"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const minVolume = 1

var (
	ErrStoragePlaceOccupied = errors.New("storage place is occupied")
	ErrOrderDoesNotFit      = errors.New("order does not fit storage place")
	ErrStoragePlaceEmpty    = errors.New("storage place is empty")
	ErrOrderIDMismatch      = errors.New("order ID does not match stored order")
)

type StoragePlace struct {
	base *ddd.BaseEntity[uuid.UUID]

	name        string
	totalVolume int
	orderID     *uuid.UUID
}

func NewStoragePlace(name string, totalVolume int) (*StoragePlace, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errs.NewValueIsRequired("name")
	}
	if totalVolume <= 0 {
		return nil, errs.NewValueMustBeGreaterOrEqual("totalVolume", totalVolume, minVolume)
	}

	return &StoragePlace{
		base:        ddd.NewBaseEntity(uuid.New()),
		name:        name,
		totalVolume: totalVolume,
	}, nil
}

func RestoreStoragePlace(id uuid.UUID, name string, totalVolume int, orderID *uuid.UUID) *StoragePlace {
	var restoredOrderID *uuid.UUID
	if orderID != nil {
		orderIDCopy := *orderID
		restoredOrderID = &orderIDCopy
	}

	return &StoragePlace{
		base:        ddd.NewBaseEntity(id),
		name:        name,
		totalVolume: totalVolume,
		orderID:     restoredOrderID,
	}
}

func (s *StoragePlace) Equals(other *StoragePlace) bool {
	return other != nil && s.base.Equal(other.base)
}

func (s *StoragePlace) ID() uuid.UUID {
	return s.base.ID()
}

func (s *StoragePlace) Name() string {
	return s.name
}

func (s *StoragePlace) TotalVolume() int {
	return s.totalVolume
}

func (s *StoragePlace) OrderID() *uuid.UUID {
	if s.orderID == nil {
		return nil
	}

	orderID := *s.orderID
	return &orderID
}

func (s *StoragePlace) CanStore(volume int) (bool, error) {
	if volume < minVolume {
		return false, errs.NewValueMustBeGreaterOrEqual("volume", volume, minVolume)
	}

	return volume <= s.totalVolume, nil
}

func (s *StoragePlace) Store(orderID uuid.UUID, volume int) error {
	if orderID == uuid.Nil {
		return errs.NewValueIsRequired("orderID")
	}
	if s.isOccupied() {
		return fmt.Errorf("%w by order %s", ErrStoragePlaceOccupied, s.orderID.String())
	}

	canStore, err := s.CanStore(volume)
	if err != nil {
		return err
	}
	if !canStore {
		return fmt.Errorf("%w: order volume %d exceeds capacity %d", ErrOrderDoesNotFit, volume, s.totalVolume)
	}

	s.orderID = &orderID
	return nil
}

func (s *StoragePlace) Clear(orderID uuid.UUID) error {
	if orderID == uuid.Nil {
		return errs.NewValueIsRequired("orderID")
	}
	if !s.isOccupied() {
		return ErrStoragePlaceEmpty
	}
	if *s.orderID != orderID {
		return ErrOrderIDMismatch
	}

	s.orderID = nil
	return nil
}

func (s *StoragePlace) isOccupied() bool {
	return s.orderID != nil
}
