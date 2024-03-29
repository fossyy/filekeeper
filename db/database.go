package db

import (
	"fmt"
	"github.com/fossyy/filekeeper/utils"
	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

type User struct {
	UserID   uuid.UUID `gorm:"primaryKey;not null;unique"`
	Username string    `gorm:"unique;not null"`
	Email    string    `gorm:"unique;not null"`
	Password string    `gorm:"not null"`
}

type File struct {
	ID         uuid.UUID `gorm:"primaryKey;not null;unique"`
	OwnerID    uuid.UUID `gorm:"not null"`
	Name       string    `gorm:"not null"`
	Size       int       `gorm:"not null"`
	Downloaded int       `gorm:"not null;default=0"`
}

func init() {
	var err error
	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", utils.Getenv("DB_USERNAME"), utils.Getenv("DB_PASSWORD"), utils.Getenv("DB_HOST"), utils.Getenv("DB_PORT"), utils.Getenv("DB_NAME"))
	DB, err = gorm.Open(mysql.Open(connection), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	DB.AutoMigrate(User{})
	DB.AutoMigrate(File{})
}
