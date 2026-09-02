//go:build integration

package kafka

import (
	"context"
	"delivery/internal/core/domain/model/order"
	"delivery/internal/generated/queues/ordereventspb"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	kafkacontainer "github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
)

func TestOrderProducer_PublishCompleted_PublishesIntegrationEvent(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	container, err := kafkacontainer.Run(
		ctx,
		"confluentinc/confluent-local:7.5.0",
		kafkacontainer.WithClusterID("delivery-producer-test"),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Errorf("terminate Kafka container: %v", err)
		}
	})

	brokers, err := container.Brokers(ctx)
	if err != nil {
		t.Fatal(err)
	}
	topic := "order.status.changed." + uuid.NewString()

	consumerClient, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumeTopics(topic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(consumerClient.Close)

	producer, err := NewOrderProducer(brokers, topic)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := producer.Close(); err != nil {
			t.Errorf("close order producer: %v", err)
		}
	})

	orderID := uuid.New()
	if err = producer.PublishCompleted(ctx, order.NewCompletedDomainEvent(orderID)); err != nil {
		t.Fatal(err)
	}

	record := readRecord(t, consumerClient, ctx)
	if string(record.Key) != orderID.String() {
		t.Errorf("Kafka key = %q, want %q", record.Key, orderID)
	}

	var event ordereventspb.OrderCompletedIntegrationEvent
	if err = proto.Unmarshal(record.Value, &event); err != nil {
		t.Fatalf("unmarshal OrderCompletedIntegrationEvent: %v", err)
	}
	if event.GetOrderId() != orderID.String() {
		t.Errorf("event.OrderId = %q, want %q", event.GetOrderId(), orderID)
	}
}

func readRecord(t *testing.T, client *kgo.Client, ctx context.Context) *kgo.Record {
	t.Helper()
	for {
		fetches := client.PollFetches(ctx)
		if fetches.IsClientClosed() {
			t.Fatal("client is closed")
		}
		if errs := fetches.Errors(); len(errs) > 0 {
			t.Fatalf("fetch error: %v", errs[0].Err)
		}
		iter := fetches.RecordIter()
		for !iter.Done() {
			return iter.Next()
		}
	}
}
