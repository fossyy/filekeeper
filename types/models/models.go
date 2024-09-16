package models

import "github.com/google/uuid"

type User struct {
	UserID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	Username string    `gorm:"type:varchar(255);unique;not null"`
	Email    string    `gorm:"type:varchar(255);unique;not null"`
	Password string    `gorm:"type:text;not null"`
	Totp     string    `gorm:"type:varchar(255);not null"`
}

type File struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	OwnerID    uuid.UUID `gorm:"type:uuid;not null"`
	Name       string    `gorm:"type:text;not null"`
	Size       uint64    `gorm:"not null"`
	TotalChunk uint64    `gorm:"not null"`
	StartHash  string    `gorm:"type:text;not null"`
	EndHash    string    `gorm:"type:text;not null"`
	IsPrivate  bool      `gorm:"not null;default:true"`
	Type       string    `gorm:"type:varchar(5);not null;default:'doc'"`
	Downloaded uint64    `gorm:"not null;default:0"`
	Owner      *User     `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE;"`
}

type Allowance struct {
	UserID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	AllowanceByte uint64    `gorm:"not null"`
	AllowanceFile uint64    `gorm:"not null"`
	User          *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
}
