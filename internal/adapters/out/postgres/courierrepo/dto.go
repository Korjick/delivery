package courierrepo

import "github.com/google/uuid"

const (
	couriersTable    = "couriers"
	assignmentsTable = "assignments"
)

type CourierDTO struct {
	ID            uuid.UUID          `gorm:"type:uuid;primaryKey"`
	Name          string             `gorm:"type:varchar(255);not null"`
	Speed         int                `gorm:"not null"`
	LocationX     int                `gorm:"not null"`
	LocationY     int                `gorm:"not null"`
	StoragePlaces []*StoragePlaceDTO `gorm:"foreignKey:CourierID;constraint:OnDelete:CASCADE;"`
}

type StoragePlaceDTO struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	CourierID   uuid.UUID  `gorm:"type:uuid;index;not null"`
	Name        string     `gorm:"type:varchar(255);not null"`
	TotalVolume int        `gorm:"not null"`
	OrderID     *uuid.UUID `gorm:"type:uuid;index"`
}

func (CourierDTO) TableName() string {
	return couriersTable
}

func (StoragePlaceDTO) TableName() string {
	return assignmentsTable
}
