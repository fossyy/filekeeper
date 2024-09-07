package service

import (
	"context"
	"encoding/json"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/types"
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

func (r *Service) GetUser(ctx context.Context, email string) (*types.UserWithExpired, error) {
	userJSON, err := app.Server.Cache.GetCache(ctx, "UserCache:"+email)
	if err == redis.Nil {
		userData, err := r.db.GetUser(email)
		if err != nil {
			return nil, err
		}

		user := &types.UserWithExpired{
			UserID:   userData.UserID,
			Username: userData.Username,
			Email:    userData.Email,
			Password: userData.Password,
			Totp:     userData.Totp,
			AccessAt: time.Now(),
		}

		newUserJSON, _ := json.Marshal(user)
		err = r.cache.SetCache(ctx, email, newUserJSON, time.Hour*24)
		if err != nil {
			return nil, err
		}

		return user, nil
	}
	if err != nil {
		return nil, err
	}

	var user types.UserWithExpired
	err = json.Unmarshal([]byte(userJSON), &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Service) DeleteUser(email string) {
	err := r.cache.DeleteCache(context.Background(), "UserCache:"+email)
	if err != nil {
		return
	}
}

func (r *Service) GetFile(id string) (*types.FileWithExpired, error) {
	fileJSON, err := r.cache.GetCache(context.Background(), "FileCache:"+id)
	if err == redis.Nil {
		uploadData, err := r.db.GetFile(id)
		if err != nil {
			return nil, err
		}

		fileCache := &types.FileWithExpired{
			ID:            uploadData.ID,
			OwnerID:       uploadData.OwnerID,
			Name:          uploadData.Name,
			Size:          uploadData.Size,
			Downloaded:    uploadData.Downloaded,
			UploadedByte:  uploadData.UploadedByte,
			UploadedChunk: uploadData.UploadedChunk,
			Done:          uploadData.Done,
			AccessAt:      time.Now(),
		}

		newFileJSON, _ := json.Marshal(fileCache)
		err = r.cache.SetCache(context.Background(), "FileCache:"+id, newFileJSON, time.Hour*24)
		if err != nil {
			return nil, err
		}
		return fileCache, nil
	}
	if err != nil {
		return nil, err
	}

	var fileCache types.FileWithExpired
	err = json.Unmarshal([]byte(fileJSON), &fileCache)
	if err != nil {
		return nil, err
	}
	return &fileCache, nil
}

func (r *Service) GetUserFile(name, ownerID string) (*types.FileWithExpired, error) {
	fileData, err := r.db.GetUserFile(name, ownerID)
	if err != nil {
		return nil, err
	}

	file, err := r.GetFile(fileData.ID.String())
	if err != nil {
		return nil, err
	}

	return file, nil
}
