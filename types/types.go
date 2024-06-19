package types

import (
	"github.com/google/uuid"
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

type FileInfo struct {
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	Chunk int64  `json:"chunk"`
}

type FileData struct {
	ID         string
	Name       string
	Size       string
	Downloaded int64
}
