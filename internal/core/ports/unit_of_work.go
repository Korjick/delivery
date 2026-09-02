package ports

import (
	"context"
	"delivery/internal/pkg/ddd"

	"gorm.io/gorm"
)

type UnitOfWork interface {
	Do(ctx context.Context, fn func(tx Tx) error) error
}

type Tx interface {
	DB() *gorm.DB
	AddEvent(event ddd.DomainEvent)
	Events() []ddd.DomainEvent
}
