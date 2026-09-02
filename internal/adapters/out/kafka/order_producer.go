package kafka

import (
	"context"
	"delivery/internal/core/domain/model/order"
	"delivery/internal/core/ports"
	"delivery/internal/generated/queues/ordereventspb"
	"delivery/internal/pkg/errs"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
)

var _ ports.OrderProducer = (*orderProducer)(nil)

type orderProducer struct {
	topic  string
	client *kgo.Client
}

func NewOrderProducer(brokers []string, topic string) (ports.OrderProducer, error) {
	if len(brokers) == 0 || strings.TrimSpace(brokers[0]) == "" {
		return nil, errs.NewValueIsRequired("brokers")
	}
	if strings.TrimSpace(topic) == "" {
		return nil, errs.NewValueIsRequired("topic")
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.DefaultProduceTopic(topic),
		kgo.RequiredAcks(kgo.AllISRAcks()),
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("create franz-go Kafka producer: %w", err)
	}

	return &orderProducer{
		topic:  topic,
		client: client,
	}, nil
}

func (p *orderProducer) PublishCompleted(ctx context.Context, event *order.CompletedDomainEvent) error {
	if event == nil {
		return errs.NewValueIsRequired("event")
	}
	if event.OrderID == uuid.Nil {
		return errs.NewValueIsRequired("orderID")
	}

	payload, err := proto.Marshal(&ordereventspb.OrderCompletedIntegrationEvent{
		OrderId: event.OrderID.String(),
	})
	if err != nil {
		return fmt.Errorf("marshal OrderCompletedIntegrationEvent: %w", err)
	}

	record := &kgo.Record{
		Topic: p.topic,
		Key:   []byte(event.OrderID.String()),
		Value: payload,
	}

	if err = p.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return fmt.Errorf("deliver OrderCompletedIntegrationEvent: %w", err)
	}

	return nil
}

func (p *orderProducer) Close() error {
	if p == nil || p.client == nil {
		return nil
	}
	p.client.Close()
	return nil
}
