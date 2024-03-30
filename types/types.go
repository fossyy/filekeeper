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
	Authenticated bool
}

type FileInfo struct {
	Name  string `json:"name"`
	Size  int    `json:"size"`
	Chunk int    `json:"chunk"`
}

type FileInfoUploaded struct {
	Name          string `json:"name"`
	Size          int    `json:"size"`
	Chunk         int    `json:"chunk"`
	UploadedChunk int    `json:"uploaded_chunk"`
}

type FileData struct {
	ID         string
	Name       string
	Size       string
	Downloaded int
}
