package product

import (
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Name      string  `gorm:"type:varchar(100);not null"`
	Price     float64 `gorm:"type:decimal(10,2);not null"`
	Stock     int     `gorm:"type:int;not null;default:0"`
	Version   int     `gorm:"type:int;not null;default:0"`
	StartTime int64   `gorm:"type:bigint;not null"`
	EndTime   int64   `gorm:"type:bigint;not null"`
}
