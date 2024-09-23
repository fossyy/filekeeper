package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/types/models"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"time"
)

type Service struct {
	db    types.Database
	cache types.CachingServer
}

func NewService(db types.Database, cache types.CachingServer) *Service {
	return &Service{
		db:    db,
		cache: cache,
	}
}

func (r *Service) GetUser(ctx context.Context, email string) (*models.User, error) {
	userJSON, err := app.Server.Cache.GetCache(ctx, "UserCache:"+email)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			userData, err := r.db.GetUser(email)
			if err != nil {
				return nil, err
			}

			user := &models.User{
				UserID:   userData.UserID,
				Username: userData.Username,
				Email:    userData.Email,
				Password: userData.Password,
				Totp:     userData.Totp,
			}

			newUserJSON, _ := json.Marshal(user)
			err = r.cache.SetCache(ctx, "UserCache:"+email, newUserJSON, time.Hour*12)
			if err != nil {
				return nil, err
			}

			return user, nil
		}
		return nil, err
	}

	var user models.User
	err = json.Unmarshal([]byte(userJSON), &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Service) DeleteUser(ctx context.Context, email string) error {
	err := r.cache.DeleteCache(ctx, "UserCache:"+email)
	if err != nil {
		return err
	}
	return nil
}

func (r *Service) GetUserStorageUsage(ctx context.Context, ownerID string) (uint64, error) {
	// TODO: Implement GetUserStorageUsage Cache
	files, err := app.Server.Database.GetFiles(ownerID, "", types.All)
	if err != nil {
		return 0, err
	}
	var total uint64 = 0
	for _, file := range files {
		total += file.Size
	}
	return total, nil
}

func (r *Service) GetFile(ctx context.Context, id string) (*models.File, error) {
	fileJSON, err := r.cache.GetCache(ctx, "FileCache:"+id)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			uploadData, err := r.db.GetFile(id)
			if err != nil {
				return nil, err
			}

			newFileJSON, _ := json.Marshal(uploadData)
			err = r.cache.SetCache(ctx, "FileCache:"+id, newFileJSON, time.Hour*24)
			if err != nil {
				return nil, err
			}
			return uploadData, nil
		}
		return nil, err
	}

	var fileCache models.File
	err = json.Unmarshal([]byte(fileJSON), &fileCache)
	if err != nil {
		return nil, err
	}
	return &fileCache, nil
}

func (r *Service) DeleteFileCache(ctx context.Context, id string) error {
	err := r.cache.DeleteCache(ctx, "FileCache:"+id)
	if err != nil {
		return err
	}
	return nil
}

func (r *Service) GetFileChunks(ctx context.Context, fileID uuid.UUID, ownerID uuid.UUID, totalChunk uint64) (*types.FileState, error) {
	fileJSON, err := r.cache.GetCache(ctx, "FileChunkCache:"+fileID.String())
	if err != nil {
		if errors.Is(err, redis.Nil) {
			prefix := fmt.Sprintf("%s/%s/chunk_", ownerID.String(), fileID.String())

			existingChunks, err := app.Server.Storage.ListObjects(ctx, prefix)
			if err != nil {
				return nil, err
			}
			missingChunk := len(existingChunks) != int(totalChunk)

			newChunkCache := types.FileState{
				Done:  !missingChunk,
				Chunk: make(map[string]bool),
			}
			for i := 0; i < int(totalChunk); i++ {
				newChunkCache.Chunk[fmt.Sprintf("chunk_%d", i)] = false
			}

			for _, chunkFile := range existingChunks {
				var chunkIndex int
				fmt.Sscanf(chunkFile, "chunk_%d", &chunkIndex)
				newChunkCache.Chunk[fmt.Sprintf("chunk_%d", chunkIndex)] = true
			}
			newChunkCacheJSON, err := json.Marshal(newChunkCache)
			if err != nil {
				return nil, err
			}
			err = r.cache.SetCache(ctx, "FileChunkCache:"+fileID.String(), newChunkCacheJSON, time.Minute*30)
			if err != nil {
				return nil, err
			}
			return &newChunkCache, nil
		}
		return nil, err
	}

	var existingCache types.FileState
	err = json.Unmarshal([]byte(fileJSON), &existingCache)
	if err != nil {
		return nil, err
	}
	return &existingCache, nil
}

func (r *Service) UpdateFileChunk(ctx context.Context, fileID uuid.UUID, ownerID uuid.UUID, chunk string, totalChunk uint64) error {
	chunks, err := r.GetFileChunks(ctx, fileID, ownerID, totalChunk)
	if err != nil {
		return err
	}

	chunks.Chunk[fmt.Sprintf("chunk_%s", chunk)] = true
	chunks.Done = true

	for i := 0; i < int(totalChunk); i++ {
		if !chunks.Chunk[fmt.Sprintf("chunk_%d", i)] {
			fmt.Println("chunk", i, " ", chunks.Chunk[fmt.Sprintf("chunk_%d", i)])
			chunks.Done = false
			break
		}
	}

	updatedChunkCacheJSON, err := json.Marshal(chunks)
	if err != nil {
		return err
	}
	err = r.cache.SetCache(ctx, "FileChunkCache:"+fileID.String(), updatedChunkCacheJSON, time.Minute*30)
	if err != nil {
		return err
	}

	return nil
}

func (r *Service) GetUserFile(ctx context.Context, fileID uuid.UUID) (*types.FileData, error) {
	fileData, err := r.GetFile(ctx, fileID.String())
	if err != nil {
		return nil, err
	}

	chunks, err := r.GetFileChunks(ctx, fileData.ID, fileData.OwnerID, fileData.TotalChunk)
	if err != nil {
		return nil, err
	}

	data := &types.FileData{
		ID:         fileData.ID,
		OwnerID:    fileData.OwnerID,
		Name:       fileData.Name,
		Size:       fileData.Size,
		TotalChunk: fileData.TotalChunk,
		StartHash:  fileData.StartHash,
		EndHash:    fileData.EndHash,
		Downloaded: fileData.Downloaded,
		IsPrivate:  fileData.IsPrivate,
		Type:       fileData.Type,
		Done:       chunks.Done,
		Chunk:      chunks.Chunk,
	}

	return data, nil
}
