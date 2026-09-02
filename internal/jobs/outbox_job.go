package jobs

import (
	"context"
	"delivery/internal/core/ports"
	"delivery/internal/pkg/ddd"
	"delivery/internal/pkg/errs"
	"delivery/internal/pkg/outbox"
	"time"
)

type OutboxJob interface{ Run() error }

const outboxPublishTimeout = 10 * time.Second

type outboxJob struct {
	unitOfWork ports.UnitOfWork
	repository ports.OutboxRepository
	registry   outbox.EventRegistry
	mediatr    ddd.Mediatr
}

func NewOutboxJob(u ports.UnitOfWork, r ports.OutboxRepository, registry outbox.EventRegistry, mediatr ddd.Mediatr) (OutboxJob, error) {
	if u == nil || r == nil || registry == nil || mediatr == nil {
		return nil, errs.NewValueIsRequired("outbox dependency")
	}
	return &outboxJob{u, r, registry, mediatr}, nil
}

func (j *outboxJob) Run() error {
	ctx := context.Background()
	messages, err := j.repository.GetNotProcessed(ctx)
	if err != nil {
		return err
	}
	for _, message := range messages {
		event, err := j.registry.DecodeDomainEvent(message)
		if err != nil {
			return err
		}
		publishCtx, cancel := context.WithTimeout(ctx, outboxPublishTimeout)
		err = j.mediatr.Publish(publishCtx, event)
		cancel()
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		message.ProcessedAtUtc = &now
		if err = j.unitOfWork.Do(ctx, func(tx ports.Tx) error { return j.repository.Update(ctx, tx, message) }); err != nil {
			return err
		}
	}
	return nil
}
