package postgres

import (
	"context"
	"delivery/internal/core/ports"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var _ ports.UnitOfWork = (*unitOfWork)(nil)

type unitOfWork struct {
	db         *gorm.DB
	outboxRepo ports.OutboxRepository
}

func NewUnitOfWork(db *gorm.DB, outboxRepo ports.OutboxRepository) (ports.UnitOfWork, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	if outboxRepo == nil {
		return nil, errors.New("outbox repository is required")
	}
	return &unitOfWork{db: db, outboxRepo: outboxRepo}, nil
}

func (u *unitOfWork) Do(ctx context.Context, fn func(tx ports.Tx) error) error {
	if fn == nil {
		return errors.New("transaction function is required")
	}

	return u.db.WithContext(ctx).Transaction(func(dbTx *gorm.DB) (err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				err = fmt.Errorf("panic in transaction: %v", recovered)
			}
		}()
		tx := newGormTx(dbTx)
		if err := fn(tx); err != nil {
			return err
		}
		return u.outboxRepo.Add(ctx, tx, tx.Events())
	})
}
