package database

import (
	"full_backend_practice/pkg/config"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitMYSQL(cfg *config.MySQLConfig, log *zap.Logger) (*gorm.DB, error) {
	var err error
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		if log != nil {
			log.Error("Mysql 数据库连接失败", zap.Error(err))
		}
		return nil, err
	}
	if log != nil {
		log.Info("MySQL 初始化成功")
	}
	return db, nil
}
