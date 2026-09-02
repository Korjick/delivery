package ports

import (
	"context"
	"delivery/internal/core/domain/model/courier"

	"github.com/google/uuid"
)

type CourierRepository interface {
	Add(ctx context.Context, tx Tx, aggregate *courier.Courier) error
	Update(ctx context.Context, tx Tx, aggregate *courier.Courier) error
	Get(ctx context.Context, tx Tx, id uuid.UUID) (*courier.Courier, error)
	GetAll(ctx context.Context, tx Tx) ([]*courier.Courier, error)
	GetAllFree(ctx context.Context, tx Tx) ([]*courier.Courier, error)
}
