package user

import (
	"fmt"

	"gorm.io/gorm"

	"full_backend_practice/pkg/mq"
)

type User struct {
	gorm.Model
	Username     string `gorm:"type:varchar(50);not null;uniqueIndex"`
	PasswordHash string `gorm:"type:varchar(255);not null"`
}

type UserDBWrapper struct {
	DB *gorm.DB
}

func (m *UserDBWrapper) RegisterUser(msg mq.UserMessage) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		err := m.FindUser(msg)
		if err != nil {
			return err
		}
		return m.CreateUser(msg)
	})
}

func (m *UserDBWrapper) FindUser(msg mq.UserMessage) error {
	var user User
	err := m.DB.Model(&User{}).Where("username = ?", msg.Username).First(&user).Error
	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("username already exists")
	}
	return nil
}
func (m *UserDBWrapper) CreateUser(msg mq.UserMessage) error {
	return m.DB.Create(&User{
		Username:     msg.Username,
		PasswordHash: msg.Password,
	}).Error
}
func (m *UserDBWrapper) LoginUser(username string) (*User, error) {
	var user User
	err := m.DB.Model(&User{}).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
