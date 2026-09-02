//go:build integration

package jobs

import (
	"context"
	kafkaout "delivery/internal/adapters/out/kafka"
	"delivery/internal/adapters/out/postgres"
	"delivery/internal/adapters/out/postgres/outboxrepo"
	"delivery/internal/core/application/eventhandlers"
	"delivery/internal/core/domain/model/order"
	"delivery/internal/core/ports"
	"delivery/internal/generated/queues/ordereventspb"
	"delivery/internal/pkg/ddd"
	"delivery/internal/pkg/outbox"
	"delivery/internal/pkg/testcnts"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	kafkacontainer "github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
	postgresgorm "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestOutboxJob_PublishesEventAndMarksMessageProcessed(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	postgresContainer, dsn, err := testcnts.StartPostgresContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := postgresContainer.Terminate(context.Background()); err != nil {
			t.Errorf("terminate PostgreSQL container: %v", err)
		}
	})

	db, err := gorm.Open(postgresgorm.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(&outbox.Message{}); err != nil {
		t.Fatal(err)
	}

	outboxRepository, err := outboxrepo.NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	unitOfWork, err := postgres.NewUnitOfWork(db, outboxRepository)
	if err != nil {
		t.Fatal(err)
	}

	kafkaContainer, err := kafkacontainer.Run(
		ctx,
		"confluentinc/confluent-local:7.5.0",
		kafkacontainer.WithClusterID("delivery-outbox-job-test"),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := kafkaContainer.Terminate(context.Background()); err != nil {
			t.Errorf("terminate Kafka container: %v", err)
		}
	})
	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		t.Fatal(err)
	}
	topic := "order.status.changed." + uuid.NewString()

	adminClient, err := kgo.NewClient(kgo.SeedBrokers(brokers...))
	if err != nil {
		t.Fatal(err)
	}
	kadmClient := kadm.NewClient(adminClient)
	if _, err = kadmClient.CreateTopics(ctx, 1, 1, nil, topic); err != nil {
		t.Fatal(err)
	}
	adminClient.Close()

	consumerClient, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumeTopics(topic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(consumerClient.Close)

	producer, err := kafkaout.NewOrderProducer(brokers, topic)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := producer.Close(); err != nil {
			t.Errorf("close order producer: %v", err)
		}
	})

	registry, err := outbox.NewEventRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if err = registry.RegisterDomainEvent(reflect.TypeOf(order.CompletedDomainEvent{})); err != nil {
		t.Fatal(err)
	}
	mediatr := ddd.NewMediatr()
	publishCompleted, err := eventhandlers.NewPublishOrderCompleted(producer)
	if err != nil {
		t.Fatal(err)
	}
	mediatr.Subscribe(publishCompleted, order.NewEmptyCompletedDomainEvent())

	event := order.NewCompletedDomainEvent(uuid.New())
	if err = unitOfWork.Do(ctx, func(tx ports.Tx) error {
		tx.AddEvent(event)
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	job, err := NewOutboxJob(unitOfWork, outboxRepository, registry, mediatr)
	if err != nil {
		t.Fatal(err)
	}
	if err = job.Run(); err != nil {
		t.Fatal(err)
	}

	record := readOutboxKafkaRecord(t, consumerClient, ctx)
	var published ordereventspb.OrderCompletedIntegrationEvent
	if err = proto.Unmarshal(record.Value, &published); err != nil {
		t.Fatalf("unmarshal OrderCompletedIntegrationEvent: %v", err)
	}
	if published.GetOrderId() != event.OrderID.String() {
		t.Errorf("published OrderId = %q, want %q", published.GetOrderId(), event.OrderID)
	}

	var stored outbox.Message
	if err = db.First(&stored, event.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.ProcessedAtUtc == nil {
		t.Error("outbox message was not marked as processed")
	}
}

func readOutboxKafkaRecord(t *testing.T, client *kgo.Client, ctx context.Context) *kgo.Record {
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
