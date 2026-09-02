package postgres

import (
	"delivery/internal/core/ports"
	"delivery/internal/pkg/ddd"

	"gorm.io/gorm"
)

var _ ports.Tx = (*gormTx)(nil)

type gormTx struct {
	db     *gorm.DB
	events []ddd.DomainEvent
}

func newGormTx(db *gorm.DB) *gormTx {
	return &gormTx{db: db, events: make([]ddd.DomainEvent, 0)}
}

func (t *gormTx) DB() *gorm.DB {
	return t.db
}

func (t *gormTx) AddEvent(event ddd.DomainEvent) {
	t.events = append(t.events, event)
}

func (t *gormTx) Events() []ddd.DomainEvent {
	return append([]ddd.DomainEvent(nil), t.events...)
}
