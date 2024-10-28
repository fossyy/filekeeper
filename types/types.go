package types

import (
	"context"
	"time"

	"github.com/fossyy/filekeeper/types/models"
	"github.com/google/uuid"
)

type FileStatus string

const (
	All     FileStatus = "all"
	Private FileStatus = "private"
	Public  FileStatus = "public"
)

type Message struct {
	Code    int
	Message string
}

type Number interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64
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
	ID         uuid.UUID
	OwnerID    uuid.UUID
	Name       string
	Size       uint64
	TotalChunk uint64
	StartHash  string
	EndHash    string
	Downloaded uint64
	IsPrivate  bool
	Type       string
	Done       bool
	Chunk      map[string]bool
}

type FileState struct {
	Done  bool
	Chunk map[string]bool
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
	GetFiles(ownerID string, query string, status FileStatus) ([]*models.File, error)
	IncrementDownloadCount(fileID string) error
	ChangeFileVisibility(fileID string) error

	InitializeTotp(email string, secret string) error
}

type CachingServer interface {
	GetCache(ctx context.Context, key string) (string, error)
	SetCache(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	DeleteCache(ctx context.Context, key string) error
	GetKeys(ctx context.Context, pattern string) ([]string, error)
	GetUser(ctx context.Context, email string) (*models.User, error)
	RemoveUserCache(ctx context.Context, email string) error
	GetFile(ctx context.Context, id string) (*models.File, error)
	GetUserFiles(ctx context.Context, ownerID uuid.UUID) ([]*models.File, error)
	RemoveUserFilesCache(ctx context.Context, ownerID uuid.UUID) error
	RemoveFileCache(ctx context.Context, id string) error
	GetFileDetail(ctx context.Context, fileID uuid.UUID) (*FileData, error)
	CalculateUserStorageUsage(ctx context.Context, ownerID string) (uint64, error)
	GetFileChunks(ctx context.Context, fileID uuid.UUID, ownerID uuid.UUID, totalChunk uint64) (*FileState, error)
	UpdateFileChunk(ctx context.Context, fileID uuid.UUID, ownerID uuid.UUID, chunk string, totalChunk uint64) error
}

type Storage interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Add(ctx context.Context, key string, data []byte) error
	DeleteRecursive(ctx context.Context, key string) error
	Delete(ctx context.Context, key string) error
	ListObjects(ctx context.Context, prefix string) ([]string, error)
}
