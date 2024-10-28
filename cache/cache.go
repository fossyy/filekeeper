package cache

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

type RedisServer struct {
	client   *redis.Client
	database types.Database
}

const (
	UserCacheKey      = "UserCache:%s"
	UserFilesCacheKey = "UserFilesCache:%s"
	FileCacheKey      = "FileCache:%s"
	FileChunkCacheKey = "FileChunkCache:%s"
)

func NewRedisServer(host, port, password string, db types.Database) types.CachingServer {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       0,
	})
	return &RedisServer{client: client, database: db}
}

func (r *RedisServer) GetCache(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r *RedisServer) SetCache(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisServer) DeleteCache(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisServer) GetKeys(ctx context.Context, pattern string) ([]string, error) {
	var cursor uint64
	var keys []string
	for {
		var newKeys []string
		var err error

		newKeys, cursor, err = r.client.Scan(ctx, cursor, pattern, 0).Result()
		if err != nil {
			return nil, err
		}

		keys = append(keys, newKeys...)

		if cursor == 0 {
			break
		}
	}
	return keys, nil
}

func (r *RedisServer) GetUser(ctx context.Context, email string) (*models.User, error) {
	cacheKey := fmt.Sprintf(UserCacheKey, email)
	userJSON, err := r.GetCache(ctx, cacheKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			userData, err := app.Server.Database.GetUser(email)
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
			err = r.SetCache(ctx, cacheKey, newUserJSON, time.Hour*12)
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

func (r *RedisServer) RemoveUserCache(ctx context.Context, email string) error {
	cacheKey := fmt.Sprintf(UserCacheKey, email)
	return r.DeleteCache(ctx, cacheKey)
}

func (r *RedisServer) GetFile(ctx context.Context, id string) (*models.File, error) {
	cacheKey := fmt.Sprintf(FileCacheKey, id)
	fileJSON, err := r.GetCache(ctx, cacheKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			fileData, err := app.Server.Database.GetFile(id)
			if err != nil {
				return nil, err
			}

			newFileJSON, _ := json.Marshal(fileData)
			err = r.SetCache(ctx, cacheKey, newFileJSON, time.Hour*24)
			if err != nil {
				return nil, err
			}
			return fileData, nil
		}
		return nil, err
	}

	var file models.File
	err = json.Unmarshal([]byte(fileJSON), &file)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *RedisServer) RemoveFileCache(ctx context.Context, id string) error {
	cacheKey := fmt.Sprintf(FileCacheKey, id)
	return r.DeleteCache(ctx, cacheKey)
}

func (r *RedisServer) GetUserFiles(ctx context.Context, ownerID uuid.UUID) ([]*models.File, error) {
	cacheKey := fmt.Sprintf(UserFilesCacheKey, ownerID.String())
	filesJSON, err := r.GetCache(ctx, cacheKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			files, err := app.Server.Database.GetFiles(ownerID.String(), "", types.All)
			if err != nil {
				return nil, err
			}

			filesJSON, err := json.Marshal(files)
			if err != nil {
				return nil, err
			}

			err = r.SetCache(ctx, cacheKey, filesJSON, time.Hour*6)
			if err != nil {
				return nil, err
			}
			return files, nil
		}
		return nil, err
	}

	var files []*models.File
	err = json.Unmarshal([]byte(filesJSON), &files)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (r *RedisServer) RemoveUserFilesCache(ctx context.Context, ownerID uuid.UUID) error {
	cacheKey := fmt.Sprintf(UserFilesCacheKey, ownerID.String())
	return r.DeleteCache(ctx, cacheKey)
}

func (r *RedisServer) CalculateUserStorageUsage(ctx context.Context, ownerID string) (uint64, error) {
	files, err := app.Server.Database.GetFiles(ownerID, "", types.All)
	if err != nil {
		return 0, err
	}

	var total uint64
	for _, file := range files {
		total += file.Size
	}

	return total, nil
}

func (r *RedisServer) GetFileChunks(ctx context.Context, fileID, ownerID uuid.UUID, totalChunk uint64) (*types.FileState, error) {
	cacheKey := fmt.Sprintf(FileChunkCacheKey, fileID.String())
	fileJSON, err := r.GetCache(ctx, cacheKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			prefix := fmt.Sprintf("%s/%s/chunk_", ownerID.String(), fileID.String())
			existingChunks, err := app.Server.Storage.ListObjects(ctx, prefix)
			if err != nil {
				return nil, err
			}

			fileState := types.FileState{
				Done:  len(existingChunks) == int(totalChunk),
				Chunk: make(map[string]bool),
			}

			for i := 0; i < int(totalChunk); i++ {
				fileState.Chunk[fmt.Sprintf("chunk_%d", i)] = false
			}

			for _, chunkFile := range existingChunks {
				var chunkIndex int
				fmt.Sscanf(chunkFile, "chunk_%d", &chunkIndex)
				fileState.Chunk[fmt.Sprintf("chunk_%d", chunkIndex)] = true
			}

			newChunkCacheJSON, err := json.Marshal(fileState)
			if err != nil {
				return nil, err
			}
			err = r.SetCache(ctx, cacheKey, newChunkCacheJSON, time.Minute*30)
			if err != nil {
				return nil, err
			}

			return &fileState, nil
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

func (r *RedisServer) UpdateFileChunk(ctx context.Context, fileID, ownerID uuid.UUID, chunk string, totalChunk uint64) error {
	chunks, err := r.GetFileChunks(ctx, fileID, ownerID, totalChunk)
	if err != nil {
		return err
	}

	chunks.Chunk[fmt.Sprintf("chunk_%s", chunk)] = true
	chunks.Done = true

	for i := 0; i < int(totalChunk); i++ {
		if !chunks.Chunk[fmt.Sprintf("chunk_%d", i)] {
			chunks.Done = false
			break
		}
	}

	updatedChunkCacheJSON, err := json.Marshal(chunks)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf(FileChunkCacheKey, fileID.String())
	err = r.SetCache(ctx, cacheKey, updatedChunkCacheJSON, time.Minute*30)
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisServer) GetFileDetail(ctx context.Context, fileID uuid.UUID) (*types.FileData, error) {
	fileData, err := r.GetFile(ctx, fileID.String())
	if err != nil {
		return nil, err
	}

	chunks, err := r.GetFileChunks(ctx, fileData.ID, fileData.OwnerID, fileData.TotalChunk)
	if err != nil {
		return nil, err
	}

	return &types.FileData{
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
	}, nil
}
