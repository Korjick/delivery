package eventhandlers

import (
	"context"
	"delivery/internal/core/domain/model/order"
	"testing"

	"github.com/google/uuid"
)

func TestPublishOrderCompleted_Handle(t *testing.T) {
	orderID := uuid.New()
	producer := &recordingOrderProducer{}
	handler, err := NewPublishOrderCompleted(producer)
	if err != nil {
		t.Fatal(err)
	}

	if err = handler.Handle(context.Background(), order.NewCompletedDomainEvent(orderID)); err != nil {
		t.Fatal(err)
	}
	if producer.event == nil || producer.event.OrderID != orderID {
		t.Errorf("published event = %#v, want completed event for %s", producer.event, orderID)
	}
}

type recordingOrderProducer struct {
	event *order.CompletedDomainEvent
}

func (p *recordingOrderProducer) PublishCompleted(_ context.Context, event *order.CompletedDomainEvent) error {
	p.event = event
	return nil
}

func (*recordingOrderProducer) Close() error { return nil }
