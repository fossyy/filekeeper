package types

import (
	"context"
	"github.com/fossyy/filekeeper/types/models"
	"github.com/google/uuid"
	"time"
)

type Message struct {
	Code    int
	Message string
}

type User struct {
	UserID        uuid.UUID
	Email         string
	Username      string
	Totp          string
	Authenticated bool
}

type Allowance struct {
	AllowanceByte        string
	AllowanceUsedByte    string
	AllowanceUsedPercent string
}

type FileData struct {
	ID         string
	Name       string
	Size       string
	IsPrivate  bool
	Type       string
	Done       bool
	Downloaded string
}

type FileWithDetail struct {
	ID         uuid.UUID
	OwnerID    uuid.UUID
	Name       string
	Size       uint64
	Downloaded uint64
	Chunk      map[string]bool
	Done       bool
}

type Database interface {
	IsUserRegistered(email string, username string) bool
	IsEmailRegistered(email string) bool

	CreateUser(user *models.User) error
	GetUser(email string) (*models.User, error)
	GetAllUsers() ([]models.User, error)
	UpdateUserPassword(email string, password string) error

	CreateAllowance(userID uuid.UUID) error
	GetAllowance(userID uuid.UUID) (*models.Allowance, error)

	CreateFile(file *models.File) error
	GetFile(fileID string) (*models.File, error)
	RenameFile(fileID string, name string) (*models.File, error)
	DeleteFile(fileID string) error
	GetUserFile(name string, ownerID string) (*models.File, error)
	GetFiles(ownerID string) ([]*models.File, error)
	IncrementDownloadCount(fileID string) error
	ChangeFileVisibility(fileID string) error

	InitializeTotp(email string, secret string) error
}

type CachingServer interface {
	GetCache(ctx context.Context, key string) (string, error)
	SetCache(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	DeleteCache(ctx context.Context, key string) error
	GetKeys(ctx context.Context, pattern string) ([]string, error)
}

type Services interface {
	GetUser(ctx context.Context, email string) (*models.User, error)
	DeleteUser(email string)
	GetFile(id string) (*models.File, error)
	GetUserFile(name, ownerID string) (*FileWithDetail, error)
	GetUserStorageUsage(ownerID string) (uint64, error)
}
