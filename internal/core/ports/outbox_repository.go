package ports

import (
	"context"
	"delivery/internal/pkg/ddd"
	"delivery/internal/pkg/outbox"
)

type OutboxRepository interface {
	Add(ctx context.Context, tx Tx, events []ddd.DomainEvent) error
	Update(ctx context.Context, tx Tx, message *outbox.Message) error
	GetNotProcessed(ctx context.Context) ([]*outbox.Message, error)
}
