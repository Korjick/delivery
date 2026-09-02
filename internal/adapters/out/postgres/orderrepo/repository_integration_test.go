//go:build integration

package orderrepo

import (
	"context"
	"delivery/internal/adapters/out/postgres"
	"delivery/internal/adapters/out/postgres/outboxrepo"
	"delivery/internal/core/domain/kernel"
	"delivery/internal/core/domain/model/order"
	"delivery/internal/core/ports"
	"delivery/internal/pkg/errs"
	"delivery/internal/pkg/outbox"
	"delivery/internal/pkg/testcnts"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	postgresgorm "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestOrderRepository_PersistsAndQueriesOrders(t *testing.T) {
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
	if err = db.AutoMigrate(&OrderDTO{}, &outbox.Message{}); err != nil {
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
	repository := NewRepository()

	order1, err := order.NewOrder(uuid.New(), kernel.MustNewLocation(3, 4), 5)
	if err != nil {
		t.Fatal(err)
	}
	order2, err := order.NewOrder(uuid.New(), kernel.MustNewLocation(6, 7), 2)
	if err != nil {
		t.Fatal(err)
	}

	// 1. Test Add
	if err = unitOfWork.Do(ctx, func(tx ports.Tx) error {
		if err := repository.Add(ctx, tx, order1); err != nil {
			return err
		}
		return repository.Add(ctx, tx, order2)
	}); err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// 2. Test Get
	var restored *order.Order
	if err = unitOfWork.Do(ctx, func(tx ports.Tx) error {
		var getErr error
		restored, getErr = repository.Get(ctx, tx, order1.ID())
		return getErr
	}); err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if !restored.Equals(order1) || restored.Volume() != 5 || restored.Status() != order.StatusCreated {
		t.Errorf("Get() returned incorrect order: %#v", restored)
	}

	// 3. Test Get not found
	if err = unitOfWork.Do(ctx, func(tx ports.Tx) error {
		_, getErr := repository.Get(ctx, tx, uuid.New())
		if !errors.Is(getErr, errs.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", getErr)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	// 4. Test GetAllCreated
	var allCreated []*order.Order
	if err = unitOfWork.Do(ctx, func(tx ports.Tx) error {
		var getErr error
		allCreated, getErr = repository.GetAllCreated(ctx, tx)
		return getErr
	}); err != nil {
		t.Fatalf("GetAllCreated() failed: %v", err)
	}
	if len(allCreated) != 2 {
		t.Errorf("GetAllCreated() returned %d orders, want 2", len(allCreated))
	}

	// 5. Test Update (Assign)
	courierID := uuid.New()
	if err = order1.Assign(courierID); err != nil {
		t.Fatal(err)
	}
	if err = unitOfWork.Do(ctx, func(tx ports.Tx) error {
		return repository.Update(ctx, tx, order1)
	}); err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	// 6. Test GetAllAssigned
	var assignedOrders []*order.Order
	if err = unitOfWork.Do(ctx, func(tx ports.Tx) error {
		var getErr error
		assignedOrders, getErr = repository.GetAllAssigned(ctx, tx)
		return getErr
	}); err != nil {
		t.Fatalf("GetAllAssigned() failed: %v", err)
	}
	if len(assignedOrders) != 1 || !assignedOrders[0].Equals(order1) {
		t.Errorf("GetAllAssigned() = %#v, want order1", assignedOrders)
	}

	// 7. Test GetAllNotCompleted (should return both order1 assigned and order2 created)
	var notCompleted []*order.Order
	if err = unitOfWork.Do(ctx, func(tx ports.Tx) error {
		var getErr error
		notCompleted, getErr = repository.GetAllNotCompleted(ctx, tx)
		return getErr
	}); err != nil {
		t.Fatalf("GetAllNotCompleted() failed: %v", err)
	}
	if len(notCompleted) != 2 {
		t.Errorf("GetAllNotCompleted() count = %d, want 2", len(notCompleted))
	}

	// 8. Test Complete & Outbox event generation
	if err = order1.Complete(); err != nil {
		t.Fatal(err)
	}
	if err = unitOfWork.Do(ctx, func(tx ports.Tx) error {
		return repository.Update(ctx, tx, order1)
	}); err != nil {
		t.Fatalf("Update() completing order failed: %v", err)
	}

	// Order1 is now completed, so GetAllNotCompleted should only return order2
	if err = unitOfWork.Do(ctx, func(tx ports.Tx) error {
		var getErr error
		notCompleted, getErr = repository.GetAllNotCompleted(ctx, tx)
		return getErr
	}); err != nil {
		t.Fatal(err)
	}
	if len(notCompleted) != 1 || !notCompleted[0].Equals(order2) {
		t.Errorf("GetAllNotCompleted() after completion = %#v, want only order2", notCompleted)
	}

	// Verify outbox message was written when order was completed
	outboxMessages, err := outboxRepository.GetNotProcessed(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(outboxMessages) == 0 {
		t.Error("expected outbox message for CompletedDomainEvent, got none")
	}
}
