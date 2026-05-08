package product

import (
	"full_backend_practice/infrastructure/mq"

	"gorm.io/gorm"
)

type productDBWrapper struct {
	DB *gorm.DB
}

func NewProductMysql(db *gorm.DB) ProductDatabase {
	_ = db.AutoMigrate(&Product{})
	return &productDBWrapper{
		DB: db,
	}
}

type Product struct {
	gorm.Model
	Name         string  `gorm:"type:varchar(100);not null"`
	Price        float64 `gorm:"type:decimal(10,2);not null"`
	SeckillPrice float64 `gorm:"type:decimal(10,2);not null"`
	Stock        int     `gorm:"type:int;not null;default:0"`
	Version      int     `gorm:"type:int;not null;default:0"`
	StartTime    int64   `gorm:"type:bigint;not null"`
	EndTime      int64   `gorm:"type:bigint;not null"`
}

func (p *productDBWrapper) GetProductList(msg mq.ProductMessage) ([]Product, error) {
	var products []Product
	err := p.DB.Find(&products).Error
	return products, err
}

func (p *productDBWrapper) GetProduct(msg mq.ProductMessage) (Product, error) {
	var product Product
	err := p.DB.Where("id = ?", msg.ProductID).First(&product).Error
	return product, err
}
