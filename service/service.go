package service

import (
	"context"
	"encoding/json"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/types/models"
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
	if err == redis.Nil {
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
		err = r.cache.SetCache(ctx, email, newUserJSON, time.Hour*24)
		if err != nil {
			return nil, err
		}

		return user, nil
	}
	if err != nil {
		return nil, err
	}

	var user models.User
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

func (r *Service) GetFile(id string) (*models.File, error) {
	fileJSON, err := r.cache.GetCache(context.Background(), "FileCache:"+id)
	if err == redis.Nil {
		uploadData, err := r.db.GetFile(id)
		if err != nil {
			return nil, err
		}

		newFileJSON, _ := json.Marshal(uploadData)
		err = r.cache.SetCache(context.Background(), "FileCache:"+id, newFileJSON, time.Hour*24)
		if err != nil {
			return nil, err
		}
		return uploadData, nil
	}
	if err != nil {
		return nil, err
	}

	var fileCache models.File
	err = json.Unmarshal([]byte(fileJSON), &fileCache)
	if err != nil {
		return nil, err
	}
	return &fileCache, nil
}

func (r *Service) GetUserFile(name, ownerID string) (*types.FileWithDetail, error) {
	fileData, err := r.db.GetUserFile(name, ownerID)
	if err != nil {
		return nil, err
	}

	dada := &types.FileWithDetail{
		ID:         fileData.ID,
		OwnerID:    fileData.OwnerID,
		Name:       fileData.Name,
		Size:       fileData.Size,
		Downloaded: fileData.Downloaded,
	}
	return dada, nil
}
