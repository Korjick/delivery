package services

import (
	"delivery/internal/core/domain/model/courier"
	"delivery/internal/core/domain/model/order"
	"delivery/internal/pkg/errs"
	"errors"
)

var (
	ErrOrderIsNotCreated      = errors.New("only created order can be dispatched")
	ErrNoSuitableCourierFound = errors.New("no suitable courier found")
)

type OrderDispatcher interface {
	Dispatch(deliveryOrder *order.Order, couriers []*courier.Courier) (*courier.Courier, error)
}

type orderDispatcher struct{}

func NewOrderDispatcher() OrderDispatcher {
	return &orderDispatcher{}
}

func (d *orderDispatcher) Dispatch(deliveryOrder *order.Order, couriers []*courier.Courier) (*courier.Courier, error) {
	if deliveryOrder == nil {
		return nil, errs.NewValueIsRequired("order")
	}
	if deliveryOrder.Status() != order.StatusCreated {
		return nil, ErrOrderIsNotCreated
	}

	var winner *courier.Courier
	var bestTime float64
	for _, candidate := range couriers {
		if candidate == nil {
			continue
		}

		canTake, err := candidate.CanTakeOrder(deliveryOrder)
		if err != nil {
			return nil, err
		}
		if !canTake {
			continue
		}

		deliveryTime, err := candidate.CalculateTimeToOrder(deliveryOrder.Location())
		if err != nil {
			return nil, err
		}
		if winner == nil || deliveryTime < bestTime {
			winner = candidate
			bestTime = deliveryTime
		}
	}
	if winner == nil {
		return nil, ErrNoSuitableCourierFound
	}
	if err := winner.TakeOrder(deliveryOrder); err != nil {
		return nil, err
	}

	return winner, nil
}
