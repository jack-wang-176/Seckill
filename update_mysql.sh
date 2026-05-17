sed -i '' '/"gorm.io\/gorm"/a\
\	"time"
' infrastructure/database/mysql.go

cat << 'INNEREOF' > patch.go
package database

import (
	"full_backend_practice/pkg/config"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitMYSQL(cfg *config.MySQLConfig, log *zap.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		if log != nil {
			log.Error("Mysql 连接失败", zap.Error(err))
		}
		return nil, err
	}

	sqlDB, err := db.DB()
	if err == nil {
		// 设置连接池参数
		sqlDB.SetMaxIdleConns(50)
		sqlDB.SetMaxOpenConns(500)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	if log != nil {
		log.Info("MySQL 初始化成功")
	}
	return db, nil
}
INNEREOF
mv patch.go infrastructure/database/mysql.go
