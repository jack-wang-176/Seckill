package mysql

import (
	"full_backend_practice/pkg/database"
	"full_backend_practice/pkg/mq"

	"gorm.io/gorm"
)

func (m *MysqlWrapper) SeckillOrder(msg mq.SeckillMessage) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&database.Product{}).
			Where("id = ? AND stock > 0", msg.ProductID).
			Update("stock", gorm.Expr("stock-1"))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Create(&database.Order{
			OrderNo:   msg.OrderNo,
			UserID:    uint(msg.UserID),
			ProductID: uint(msg.ProductID),
			Status:    1,
		}).Error
	})

}
