package database

import (
	"full_backend_practice/pkg/config"
	"full_backend_practice/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitMYSQL(cfg *config.MySQLConfig) (*gorm.DB, error) {
	var err error
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		logger.Log.Fatal("Mysql 数据库连接失败", zap.Error(err))
	}
	logger.Log.Info("MySQL 初始化成功")
	return db, nil
}
