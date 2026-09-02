package eventhandlers

import (
	"context"
	"delivery/internal/core/domain/model/order"
	"delivery/internal/core/ports"
	"delivery/internal/pkg/ddd"
	"delivery/internal/pkg/errs"
	"fmt"
)

type publishOrderCompleted struct{ producer ports.OrderProducer }

func NewPublishOrderCompleted(producer ports.OrderProducer) (ddd.EventHandler, error) {
	if producer == nil {
		return nil, errs.NewValueIsRequired("orderProducer")
	}
	return &publishOrderCompleted{producer: producer}, nil
}

func (h *publishOrderCompleted) Handle(ctx context.Context, event ddd.DomainEvent) error {
	completed, ok := event.(*order.CompletedDomainEvent)
	if !ok {
		return fmt.Errorf("unexpected event: %T", event)
	}
	return h.producer.PublishCompleted(ctx, completed)
}
