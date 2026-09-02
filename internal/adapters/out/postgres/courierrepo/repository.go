package courierrepo

import (
	"context"
	"delivery/internal/core/domain/model/courier"
	"delivery/internal/core/ports"
	"delivery/internal/pkg/errs"
	"delivery/internal/pkg/must"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ ports.CourierRepository = (*Repository)(nil)

type Repository struct{}

func NewRepository() ports.CourierRepository {
	return &Repository{}
}

func (r *Repository) Add(ctx context.Context, tx ports.Tx, aggregate *courier.Courier) error {
	must.NotNil(tx, "tx")
	must.NotNil(aggregate, "aggregate")

	dto := DomainToDTO(aggregate)
	if err := tx.DB().WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(&dto).Error; err != nil {
		return err
	}
	moveEvents(tx, aggregate)
	return nil
}

func (r *Repository) Update(ctx context.Context, tx ports.Tx, aggregate *courier.Courier) error {
	must.NotNil(tx, "tx")
	must.NotNil(aggregate, "aggregate")

	dto := DomainToDTO(aggregate)
	if err := tx.DB().WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Save(&dto).Error; err != nil {
		return err
	}
	moveEvents(tx, aggregate)
	return nil
}

func (r *Repository) Get(ctx context.Context, tx ports.Tx, id uuid.UUID) (*courier.Courier, error) {
	must.NotNil(tx, "tx")

	var dto CourierDTO
	result := tx.DB().WithContext(ctx).Preload(clause.Associations).First(&dto, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errs.NewNotFound("courier", id)
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return DTOToDomain(dto), nil
}

func (r *Repository) GetAll(ctx context.Context, tx ports.Tx) ([]*courier.Courier, error) {
	must.NotNil(tx, "tx")

	var dtos []CourierDTO
	result := tx.DB().WithContext(ctx).
		Preload(clause.Associations).
		Find(&dtos)
	if result.Error != nil {
		return nil, result.Error
	}

	couriers := make([]*courier.Courier, 0, len(dtos))
	for _, dto := range dtos {
		couriers = append(couriers, DTOToDomain(dto))
	}
	return couriers, nil
}

func (r *Repository) GetAllFree(ctx context.Context, tx ports.Tx) ([]*courier.Courier, error) {
	must.NotNil(tx, "tx")

	var dtos []CourierDTO
	result := tx.DB().WithContext(ctx).
		Preload(clause.Associations).
		Where("NOT EXISTS (SELECT 1 FROM assignments WHERE assignments.courier_id = couriers.id AND assignments.order_id IS NOT NULL)").
		Find(&dtos)
	if result.Error != nil {
		return nil, result.Error
	}

	couriers := make([]*courier.Courier, 0, len(dtos))
	for _, dto := range dtos {
		couriers = append(couriers, DTOToDomain(dto))
	}
	return couriers, nil
}

func moveEvents(tx ports.Tx, aggregate *courier.Courier) {
	for _, event := range aggregate.GetDomainEvents() {
		tx.AddEvent(event)
	}
	aggregate.ClearDomainEvents()
}
