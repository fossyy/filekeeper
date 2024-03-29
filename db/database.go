package db

import (
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

	dsn := "root@tcp(127.0.0.1:3306)/filekeeper?charset=utf8mb4&parseTime=True&loc=Local"
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	DB.AutoMigrate(User{})
	DB.AutoMigrate(File{})
}
