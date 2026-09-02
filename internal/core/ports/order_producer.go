package ports

import (
	"context"
	"delivery/internal/core/domain/model/order"
)

type OrderProducer interface {
	PublishCompleted(ctx context.Context, event *order.CompletedDomainEvent) error
	Close() error
}
