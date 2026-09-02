package orderrepo

import (
	"delivery/internal/core/domain/model/order"

	"github.com/google/uuid"
)

const ordersTable = "orders"

type OrderDTO struct {
	ID        uuid.UUID    `gorm:"type:uuid;primaryKey"`
	CourierID *uuid.UUID   `gorm:"type:uuid;index"`
	LocationX int          `gorm:"not null"`
	LocationY int          `gorm:"not null"`
	Volume    int          `gorm:"not null"`
	Status    order.Status `gorm:"type:varchar(20);index;not null"`
}

func (OrderDTO) TableName() string {
	return ordersTable
}
