package database

import (
	"full_backend_practice/pkg/config"
	"full_backend_practice/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

//var DB *gorm.DB

type User struct {
	gorm.Model
	Username     string `gorm:"type:varchar(50);not null;uniqueIndex"`
	PasswordHash string `gorm:"type:varchar(255);not null"`
}

type Product struct {
	gorm.Model
	Name      string  `gorm:"type:varchar(100);not null"`
	Price     float64 `gorm:"type:decimal(10,2);not null"`
	Stock     int     `gorm:"type:int;not null;default:0"`
	Version   int     `gorm:"type:int;not null;default:0"` // 乐观锁版本号
	StartTime int64   `gorm:"type:bigint;not null"`        // 推荐用时间戳存储
	EndTime   int64   `gorm:"type:bigint;not null"`
}

type Order struct {
	gorm.Model
	OrderNo   string `gorm:"type:varchar(64);not null;uniqueIndex"`
	UserID    uint   `gorm:"not null;uniqueIndex:idx_user_product"` // 联合 唯一索引兜底防重复
	ProductID uint   `gorm:"not null;uniqueIndex:idx_user_product"`
	Status    int8   `gorm:"type:tinyint;not null;default:0"` // 0-排队中, 1-成功, 2-失败
}

func InitMYSQL(cfg *config.MySQLConfig) (*gorm.DB, error) {
	var err error
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		logger.Log.Fatal("Mysql 数据库连接失败", zap.Error(err))
	}
	err = db.AutoMigrate(&User{}, &Product{}, &Order{})
	if err != nil {
		logger.Log.Fatal("数据库表创建失败", zap.Error(err))
	}
	logger.Log.Info("MySQL 初始化且表结构同步成功")
	return db, nil
}
