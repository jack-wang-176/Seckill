package user

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"full_backend_practice/infrastructure/mq"
)

type User struct {
	gorm.Model
	Username     string `gorm:"type:varchar(50);not null;uniqueIndex"`
	PasswordHash string `gorm:"type:varchar(255);not null"`
}

type userDBWrapper struct {
	DB *gorm.DB
}

func NewUserMysql(db *gorm.DB, ctx context.Context) UserDatabase {
	_ = db.WithContext(ctx).AutoMigrate(&User{})
	return &userDBWrapper{DB: db}
}

func (m *userDBWrapper) RegisterUser(ctx context.Context, msg mq.UserMessage) error {
	return m.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := m.FindUser(ctx, msg)
		if err != nil {
			return err
		}
		return m.CreateUser(ctx, msg)
	})
}

func (m *userDBWrapper) FindUser(ctx context.Context, msg mq.UserMessage) error {
	var user User
	err := m.DB.WithContext(ctx).Model(&User{}).Where("username = ?", msg.Username).First(&user).Error
	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("username already exists")
	}
	return nil
}
func (m *userDBWrapper) CreateUser(ctx context.Context, msg mq.UserMessage) error {
	return m.DB.WithContext(ctx).Create(&User{
		Username:     msg.Username,
		PasswordHash: msg.Password,
	}).Error
}
func (m *userDBWrapper) LoginUser(ctx context.Context, username string) (*User, error) {
	var user User
	err := m.DB.WithContext(ctx).Model(&User{}).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
