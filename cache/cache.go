package cache

import (
	"fmt"
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/utils"
	"github.com/google/uuid"
	"sync"
	"time"
)

type UserWithExpired struct {
	UserID   uuid.UUID
	Username string
	Email    string
	Password string
	AccessAt time.Time
	mu       sync.Mutex
}

type FileWithExpired struct {
	ID            uuid.UUID
	OwnerID       uuid.UUID
	Name          string
	Size          int64
	Downloaded    int64
	UploadedByte  int64
	UploadedChunk int64
	Done          bool
	AccessAt      time.Time
	mu            sync.Mutex
}

var log *logger.AggregatedLogger
var userCache map[string]*UserWithExpired
var fileCache map[string]*FileWithExpired

func init() {
	log = logger.Logger()

	userCache = make(map[string]*UserWithExpired)
	fileCache = make(map[string]*FileWithExpired)
	ticker := time.NewTicker(time.Minute)

	go func() {
		for {
			<-ticker.C
			currentTime := time.Now()
			cacheClean := 0
			cleanID := utils.GenerateRandomString(10)
			log.Info(fmt.Sprintf("Cache cleanup [user] [%s] initiated at %02d:%02d:%02d", cleanID, currentTime.Hour(), currentTime.Minute(), currentTime.Second()))

			for _, user := range userCache {
				user.mu.Lock()
				if currentTime.Sub(user.AccessAt) > time.Hour*8 {
					delete(userCache, user.Email)
					cacheClean++
				}
				user.mu.Unlock()
			}

			log.Info(fmt.Sprintf("Cache cleanup [user] [%s] completed: %d entries removed. Finished at %s", cleanID, cacheClean, time.Since(currentTime)))
		}
	}()

	go func() {
		for {
			<-ticker.C
			currentTime := time.Now()
			cacheClean := 0
			cleanID := utils.GenerateRandomString(10)
			log.Info(fmt.Sprintf("Cache cleanup [files] [%s] initiated at %02d:%02d:%02d", cleanID, currentTime.Hour(), currentTime.Minute(), currentTime.Second()))

			for _, file := range fileCache {
				file.mu.Lock()
				if currentTime.Sub(file.AccessAt) > time.Minute*1 {
					db.DB.UpdateUploadedByte(file.UploadedByte, file.ID.String())
					db.DB.UpdateUploadedChunk(file.UploadedChunk, file.ID.String())
					delete(fileCache, file.ID.String())
					cacheClean++
				}
				file.mu.Unlock()
			}

			log.Info(fmt.Sprintf("Cache cleanup [files] [%s] completed: %d entries removed. Finished at %s", cleanID, cacheClean, time.Since(currentTime)))
		}
	}()
}

func GetUser(email string) (*UserWithExpired, error) {
	if user, ok := userCache[email]; ok {
		return user, nil
	}

	userData, err := db.DB.GetUser(email)
	if err != nil {
		return nil, err
	}

	userCache[email] = &UserWithExpired{
		UserID:   userData.UserID,
		Username: userData.Username,
		Email:    userData.Email,
		Password: userData.Password,
		AccessAt: time.Now(),
	}

	return userCache[email], nil
}

func DeleteUser(email string) {
	userCache[email].mu.Lock()
	defer userCache[email].mu.Unlock()

	delete(userCache, email)
}

func GetFile(id string) (*FileWithExpired, error) {
	if file, ok := fileCache[id]; ok {
		file.AccessAt = time.Now()
		return file, nil
	}

	uploadData, err := db.DB.GetFile(id)
	if err != nil {
		return nil, err
	}

	fileCache[id] = &FileWithExpired{
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

	return fileCache[id], nil
}

func (file *FileWithExpired) UpdateProgress(index int64, size int64) {
	file.UploadedChunk = index
	file.UploadedByte = size
	file.AccessAt = time.Now()
}

func GetUserFile(name, ownerID string) (*FileWithExpired, error) {
	fileData, err := db.DB.GetUserFile(name, ownerID)
	if err != nil {
		return nil, err
	}

	file, err := GetFile(fileData.ID.String())
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (file *FileWithExpired) FinalizeFileUpload() {
	db.DB.UpdateUploadedByte(file.UploadedByte, file.ID.String())
	db.DB.UpdateUploadedChunk(file.UploadedChunk, file.ID.String())
	db.DB.FinalizeFileUpload(file.ID.String())
	delete(fileCache, file.ID.String())
	return
}

//func DeleteUploadInfo(id string) {
//	filesUploadedCache[id].mu.Lock()
//	defer filesUploadedCache[id].mu.Unlock()
//
//	delete(filesUploadedCache, id)
//}
