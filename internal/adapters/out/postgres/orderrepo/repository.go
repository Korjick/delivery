package orderrepo

import (
	"context"
	"delivery/internal/core/domain/model/order"
	"delivery/internal/core/ports"
	"delivery/internal/pkg/errs"
	"delivery/internal/pkg/must"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct{}

func NewRepository() ports.OrderRepository {
	return &Repository{}
}

func (r *Repository) Add(ctx context.Context, tx ports.Tx, aggregate *order.Order) error {
	must.NotNil(tx, "tx")
	must.NotNil(aggregate, "aggregate")

	dto := DomainToDTO(aggregate)
	if err := tx.DB().WithContext(ctx).Create(&dto).Error; err != nil {
		return err
	}
	moveEvents(tx, aggregate)
	return nil
}

func (r *Repository) Update(ctx context.Context, tx ports.Tx, aggregate *order.Order) error {
	must.NotNil(tx, "tx")
	must.NotNil(aggregate, "aggregate")

	dto := DomainToDTO(aggregate)
	if err := tx.DB().WithContext(ctx).Save(&dto).Error; err != nil {
		return err
	}
	moveEvents(tx, aggregate)
	return nil
}

func (r *Repository) Get(ctx context.Context, tx ports.Tx, id uuid.UUID) (*order.Order, error) {
	must.NotNil(tx, "tx")

	var dto OrderDTO
	result := tx.DB().WithContext(ctx).First(&dto, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errs.NewNotFound("order", id)
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return DTOToDomain(dto), nil
}

func (r *Repository) GetAllCreated(ctx context.Context, tx ports.Tx) ([]*order.Order, error) {
	must.NotNil(tx, "tx")

	var dtos []OrderDTO
	result := tx.DB().WithContext(ctx).Where("status = ?", order.StatusCreated).Find(&dtos)
	if result.Error != nil {
		return nil, result.Error
	}

	orders := make([]*order.Order, 0, len(dtos))
	for _, dto := range dtos {
		orders = append(orders, DTOToDomain(dto))
	}
	return orders, nil
}

func (r *Repository) GetAllAssigned(ctx context.Context, tx ports.Tx) ([]*order.Order, error) {
	must.NotNil(tx, "tx")

	var dtos []OrderDTO
	if err := tx.DB().WithContext(ctx).Where("status = ?", order.StatusAssigned).Find(&dtos).Error; err != nil {
		return nil, err
	}
	orders := make([]*order.Order, 0, len(dtos))
	for _, dto := range dtos {
		orders = append(orders, DTOToDomain(dto))
	}
	return orders, nil
}

func (r *Repository) GetAllNotCompleted(ctx context.Context, tx ports.Tx) ([]*order.Order, error) {
	must.NotNil(tx, "tx")

	var dtos []OrderDTO
	if err := tx.DB().WithContext(ctx).
		Where("status IN ?", []order.Status{order.StatusCreated, order.StatusAssigned}).
		Find(&dtos).Error; err != nil {
		return nil, err
	}
	orders := make([]*order.Order, 0, len(dtos))
	for _, dto := range dtos {
		orders = append(orders, DTOToDomain(dto))
	}
	return orders, nil
}

func moveEvents(tx ports.Tx, aggregate *order.Order) {
	for _, event := range aggregate.GetDomainEvents() {
		tx.AddEvent(event)
	}
	aggregate.ClearDomainEvents()
}
