package models

import "github.com/google/uuid"

type User struct {
	UserID   uuid.UUID `gorm:"primaryKey;not null;unique"`
	Username string    `gorm:"unique;not null"`
	Email    string    `gorm:"unique;not null"`
	Password string    `gorm:"not null"`
	Totp     string    `gorm:"not null"`
}

type File struct {
	ID         uuid.UUID `gorm:"primaryKey;not null;unique"`
	OwnerID    uuid.UUID `gorm:"not null"`
	Name       string    `gorm:"not null"`
	Size       uint64    `gorm:"not null"`
	TotalChunk uint64    `gorm:"not null"`
	Downloaded uint64    `gorm:"not null;default=0"`
}
