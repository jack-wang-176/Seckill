package order

import (
	"full_backend_practice/pkg/mq"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	OrderNo   string  `gorm:"type:varchar(64);not null;uniqueIndex"`
	ProductID uint    `gorm:"not null;uniqueIndex:idx_user_product"`
	UserID    uint    `gorm:"not null;uniqueIndex:idx_user_product"`
	Amount    float64 `gorm:"type:decimal(10,2);not null"`
	Status    int8    `gorm:"type:tinyint;not null;default:0"`
}

type Product struct {
	ID    uint
	Stock int
}

func (Product) TableName() string {
	return "products"
}

func NewOrderMysql(db *gorm.DB) OrderDatabase {
	_ = db.AutoMigrate(&Order{})
	return &orderDBWrapper{DB: db}
}

type orderDBWrapper struct {
	DB *gorm.DB
}

func (m *orderDBWrapper) SeckillOrder(msg mq.SeckillMessage) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Product{}).
			Where("id = ? AND stock > 0", msg.ProductID).
			Update("stock", gorm.Expr("stock-1"))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Create(&Order{
			OrderNo:   msg.OrderNo,
			ProductID: uint(msg.ProductID),
			Status:    1,
		}).Error
	})

}
