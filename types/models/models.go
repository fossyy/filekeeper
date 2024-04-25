package models

import "github.com/google/uuid"

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

type FilesUploaded struct {
	UploadID uuid.UUID `gorm:"primaryKey;not null;unique"`
	FileID   uuid.UUID `gorm:"not null"`
	OwnerID  uuid.UUID `gorm:"not null"`
	Name     string    `gorm:"not null"`
	Size     int       `gorm:"not null"`
	Uploaded int       `gorm:"not null;default=0"`
	Done     bool      `gorm:"not null;default=false"`
}
