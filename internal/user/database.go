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

type UserDatabase interface {
	RegisterUser(msg mq.UserMessage) error
	FindUser(msg mq.UserMessage) error
	CreateUser(msg mq.UserMessage) error
	LoginUser(username string) (*User, error)
}
type userDBWrapper struct {
	DB *gorm.DB
}

func (m *userDBWrapper) RegisterUser(msg mq.UserMessage) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		err := m.FindUser(msg)
		if err != nil {
			return err
		}
		return m.CreateUser(msg)
	})
}

func (m *userDBWrapper) FindUser(msg mq.UserMessage) error {
	var user User
	err := m.DB.Model(&User{}).Where("username = ?", msg.Username).First(&user).Error
	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("username already exists")
	}
	return nil
}
func (m *userDBWrapper) CreateUser(msg mq.UserMessage) error {
	return m.DB.Create(&User{
		Username:     msg.Username,
		PasswordHash: msg.Password,
	}).Error
}
func (m *userDBWrapper) LoginUser(username string) (*User, error) {
	var user User
	err := m.DB.Model(&User{}).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
