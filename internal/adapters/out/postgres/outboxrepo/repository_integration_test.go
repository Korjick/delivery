//go:build integration

package outboxrepo

import (
	"context"
	"delivery/internal/adapters/out/postgres"
	"delivery/internal/core/domain/model/order"
	"delivery/internal/core/ports"
	"delivery/internal/pkg/outbox"
	"delivery/internal/pkg/testcnts"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	postgresgorm "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestRepository_PersistsAndMarksDomainEventProcessed(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx := context.Background()

	container, dsn, err := testcnts.StartPostgresContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
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

	repository, err := NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	unitOfWork, err := postgres.NewUnitOfWork(db, repository)
	if err != nil {
		t.Fatal(err)
	}

	event := order.NewCompletedDomainEvent(uuid.New())
	if err = unitOfWork.Do(ctx, func(tx ports.Tx) error {
		tx.AddEvent(event)
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	messages, err := repository.GetNotProcessed(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 {
		t.Fatalf("not processed message count = %d, want 1", len(messages))
	}
	if messages[0].ID != event.ID {
		t.Errorf("message ID = %s, want %s", messages[0].ID, event.ID)
	}

	registry, err := outbox.NewEventRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if err = registry.RegisterDomainEvent(reflect.TypeOf(order.CompletedDomainEvent{})); err != nil {
		t.Fatal(err)
	}
	decoded, err := registry.DecodeDomainEvent(messages[0])
	if err != nil {
		t.Fatal(err)
	}
	completed, ok := decoded.(*order.CompletedDomainEvent)
	if !ok {
		t.Fatalf("decoded event type = %T, want *order.CompletedDomainEvent", decoded)
	}
	if completed.OrderID != event.OrderID {
		t.Errorf("decoded OrderID = %s, want %s", completed.OrderID, event.OrderID)
	}

	processedAt := time.Now().UTC()
	messages[0].ProcessedAtUtc = &processedAt
	if err = unitOfWork.Do(ctx, func(tx ports.Tx) error {
		return repository.Update(ctx, tx, messages[0])
	}); err != nil {
		t.Fatal(err)
	}

	messages, err = repository.GetNotProcessed(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 0 {
		t.Fatalf("not processed message count after update = %d, want 0", len(messages))
	}
}
