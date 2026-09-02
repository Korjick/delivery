package outboxrepo

import (
	"context"
	"delivery/internal/core/ports"
	"delivery/internal/pkg/ddd"
	"delivery/internal/pkg/errs"
	"delivery/internal/pkg/must"
	"delivery/internal/pkg/outbox"

	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) (ports.OutboxRepository, error) {
	if db == nil {
		return nil, errs.NewValueIsRequired("db")
	}
	return &Repository{db: db}, nil
}

func (r *Repository) Add(ctx context.Context, tx ports.Tx, events []ddd.DomainEvent) error {
	must.NotNil(tx, "tx")
	if len(events) == 0 {
		return nil
	}
	messages := make([]outbox.Message, 0, len(events))
	for _, event := range events {
		message, err := outbox.EncodeDomainEvent(event)
		if err != nil {
			return err
		}
		messages = append(messages, message)
	}
	return tx.DB().WithContext(ctx).Create(&messages).Error
}

func (r *Repository) Update(ctx context.Context, tx ports.Tx, message *outbox.Message) error {
	must.NotNil(tx, "tx")
	must.NotNil(message, "message")
	return tx.DB().WithContext(ctx).Save(message).Error
}

func (r *Repository) GetNotProcessed(ctx context.Context) ([]*outbox.Message, error) {
	var messages []*outbox.Message
	if err := r.db.WithContext(ctx).
		Where("processed_at_utc IS NULL").
		Order("occurred_at_utc ASC").
		Limit(20).Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}
