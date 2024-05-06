package models

import "github.com/google/uuid"

type User struct {
	UserID   uuid.UUID `gorm:"primaryKey;not null;unique"`
	Username string    `gorm:"unique;not null"`
	Email    string    `gorm:"unique;not null"`
	Password string    `gorm:"not null"`
}

type File struct {
	ID            uuid.UUID `gorm:"primaryKey;not null;unique"`
	OwnerID       uuid.UUID `gorm:"not null"`
	Name          string    `gorm:"not null"`
	Size          int64     `gorm:"not null"`
	Downloaded    int64     `gorm:"not null;default=0"`
	UploadedByte  int64     `gorm:"not null;default=0"`
	UploadedChunk int64     `gorm:"not null;default=0"`
	Done          bool      `gorm:"not null;default=false"`
}
