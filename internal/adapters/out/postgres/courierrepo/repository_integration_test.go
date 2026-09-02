//go:build integration

package courierrepo

import (
	"context"
	"delivery/internal/adapters/out/postgres"
	"delivery/internal/adapters/out/postgres/outboxrepo"
	"delivery/internal/core/domain/kernel"
	"delivery/internal/core/domain/model/courier"
	"delivery/internal/core/ports"
	"delivery/internal/pkg/testcnts"
	"testing"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	postgresgorm "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestCourierRepository_PersistsAndRestoresCourier(t *testing.T) {
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
	if err = db.AutoMigrate(&CourierDTO{}, &StoragePlaceDTO{}); err != nil {
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

	busyCourier, err := courier.NewCourier("Busy", 2, kernel.MustNewLocation(1, 1))
	if err != nil {
		t.Fatal(err)
	}
	if err = busyCourier.AddStoragePlace("Trunk", 100); err != nil {
		t.Fatal(err)
	}
	freeCourier, err := courier.NewCourier("Free", 1, kernel.MustNewLocation(2, 2))
	if err != nil {
		t.Fatal(err)
	}

	if err = unitOfWork.Do(ctx, func(tx ports.Tx) error {
		if err := repository.Add(ctx, tx, busyCourier); err != nil {
			return err
		}
		return repository.Add(ctx, tx, freeCourier)
	}); err != nil {
		t.Fatal(err)
	}

	var restored *courier.Courier
	if err = unitOfWork.Do(ctx, func(tx ports.Tx) error {
		var getErr error
		restored, getErr = repository.Get(ctx, tx, busyCourier.ID())
		return getErr
	}); err != nil {
		t.Fatal(err)
	}
	if !restored.Equals(busyCourier) || restored.Name() != "Busy" || restored.Speed() != 2 || restored.Location() != (kernel.MustNewLocation(1, 1)) || len(restored.StoragePlaces()) != 2 {
		t.Errorf("Get() restored invalid courier: %#v", restored)
	}

	if err = busyCourier.StoragePlaces()[0].Store(uuid.New(), 1); err != nil {
		t.Fatal(err)
	}
	if err = unitOfWork.Do(ctx, func(tx ports.Tx) error {
		return repository.Update(ctx, tx, busyCourier)
	}); err != nil {
		t.Fatal(err)
	}

	var freeCouriers []*courier.Courier
	if err = unitOfWork.Do(ctx, func(tx ports.Tx) error {
		var getErr error
		freeCouriers, getErr = repository.GetAllFree(ctx, tx)
		return getErr
	}); err != nil {
		t.Fatal(err)
	}
	if len(freeCouriers) != 1 || !freeCouriers[0].Equals(freeCourier) {
		t.Errorf("GetAllFree() = %#v, want only %s", freeCouriers, freeCourier.ID())
	}

	var allCouriers []*courier.Courier
	if err = unitOfWork.Do(ctx, func(tx ports.Tx) error {
		var getErr error
		allCouriers, getErr = repository.GetAll(ctx, tx)
		return getErr
	}); err != nil {
		t.Fatal(err)
	}
	if len(allCouriers) != 2 {
		t.Errorf("GetAll() returned %d couriers, want 2", len(allCouriers))
	}
}
