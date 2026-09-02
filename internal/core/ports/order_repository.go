package ports

import (
	"context"
	"delivery/internal/core/domain/model/order"

	"github.com/google/uuid"
)

type OrderRepository interface {
	Add(ctx context.Context, tx Tx, aggregate *order.Order) error
	Update(ctx context.Context, tx Tx, aggregate *order.Order) error
	Get(ctx context.Context, tx Tx, id uuid.UUID) (*order.Order, error)
	GetAllCreated(ctx context.Context, tx Tx) ([]*order.Order, error)
	GetAllAssigned(ctx context.Context, tx Tx) ([]*order.Order, error)
	GetAllNotCompleted(ctx context.Context, tx Tx) ([]*order.Order, error)
}
