//go:build integration

package inboxrepo

import (
	"context"
	"delivery/internal/adapters/out/postgres"
	"delivery/internal/adapters/out/postgres/outboxrepo"
	"delivery/internal/core/ports"
	"delivery/internal/pkg/inbox"
	"delivery/internal/pkg/testcnts"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	postgresgorm "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestRepository_TryAdd_IsIdempotent(t *testing.T) {
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
	if err = db.AutoMigrate(&inbox.Message{}); err != nil {
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

	const messageID = "basket.confirmed:0:42"
	var firstInserted bool
	if err = unitOfWork.Do(ctx, func(tx ports.Tx) error {
		var tryAddErr error
		firstInserted, tryAddErr = repository.TryAdd(ctx, tx, messageID)
		return tryAddErr
	}); err != nil {
		t.Fatal(err)
	}
	if !firstInserted {
		t.Fatal("first TryAdd() = false, want true")
	}

	var secondInserted bool
	if err = unitOfWork.Do(ctx, func(tx ports.Tx) error {
		var tryAddErr error
		secondInserted, tryAddErr = repository.TryAdd(ctx, tx, messageID)
		return tryAddErr
	}); err != nil {
		t.Fatal(err)
	}
	if secondInserted {
		t.Fatal("second TryAdd() = true, want false")
	}

	var count int64
	if err = db.Model(&inbox.Message{}).Where("id = ?", messageID).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("stored message count = %d, want 1", count)
	}
}
