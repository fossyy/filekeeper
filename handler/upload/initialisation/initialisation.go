package initialisation

import (
	"encoding/json"
	"errors"
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/middleware"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/types/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

var log *logger.AggregatedLogger

func init() {
	log = logger.Logger()
}

func POST(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("Session")
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	storeSession, err := session.Store.Get(cookie.Value)
	if err != nil {
		if errors.Is(err, &session.SessionNotFound{}) {
			storeSession.Destroy(w)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	userSession := middleware.GetUser(storeSession)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	var fileInfo types.FileInfo
	if err := json.Unmarshal(body, &fileInfo); err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	fileData, err := getFile(fileInfo.Name, userSession.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			upload, err := handleNewUpload(userSession, fileInfo)
			if err != nil {
				handleError(w, err, http.StatusInternalServerError)
				return
			}
			respondJSON(w, upload)
			return
		}
		respondErrorJSON(w, err, http.StatusInternalServerError)
		return
	}

	info, err := GetUploadInfo(fileData.ID.String())
	if err != nil {
		log.Error(err.Error())
		return
	}

	if info.Done {
		respondJSON(w, map[string]bool{"Done": true})
		return
	}
	respondJSON(w, info)
}

func getFile(name string, ownerID uuid.UUID) (models.File, error) {
	var data models.File
	err := db.DB.Table("files").Where("name = ? AND owner_id = ?", name, ownerID).First(&data).Error
	if err != nil {
		return data, err
	}
	return data, nil
}

func handleNewUpload(user types.User, file types.FileInfo) (models.FilesUploaded, error) {
	uploadDir := "uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		log.Error(err.Error())
		err := os.Mkdir(uploadDir, os.ModePerm)
		if err != nil {
			log.Error(err.Error())
			return models.FilesUploaded{}, err
		}
	}

	fileID := uuid.New()
	ownerID := user.UserID

	currentDir, _ := os.Getwd()
	basePath := filepath.Join(currentDir, uploadDir)
	saveFolder := filepath.Join(basePath, ownerID.String(), fileID.String())
	if filepath.Dir(saveFolder) != filepath.Join(basePath, ownerID.String()) {
		return models.FilesUploaded{}, errors.New("invalid path")
	}

	err := os.MkdirAll(saveFolder, os.ModePerm)
	if err != nil {
		log.Error(err.Error())
		return models.FilesUploaded{}, err
	}

	newFile := models.File{
		ID:         fileID,
		OwnerID:    ownerID,
		Name:       file.Name,
		Size:       file.Size,
		Downloaded: 0,
	}
	err = db.DB.Create(&newFile).Error
	if err != nil {
		log.Error(err.Error())
		return models.FilesUploaded{}, err
	}

	filesUploaded := models.FilesUploaded{
		UploadID: uuid.New(),
		FileID:   fileID,
		OwnerID:  ownerID,
		Name:     file.Name,
		Size:     file.Size,
		Uploaded: -1,
		Done:     false,
	}

	err = db.DB.Create(&filesUploaded).Error
	if err != nil {
		log.Error(err.Error())
		return models.FilesUploaded{}, err
	}
	return filesUploaded, nil
}

func GetUploadInfo(fileID string) (*models.FilesUploaded, error) {
	var data *models.FilesUploaded
	err := db.DB.Table("files_uploadeds").Where("file_id = ?", fileID).First(&data).Error
	if err != nil {
		return data, err
	}
	return data, nil
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		handleError(w, err, http.StatusInternalServerError)
	}
}

func respondErrorJSON(w http.ResponseWriter, err error, statusCode int) {
	w.WriteHeader(statusCode)
	respondJSON(w, map[string]string{"error": err.Error()})
}

func handleError(w http.ResponseWriter, err error, statusCode int) {
	http.Error(w, err.Error(), statusCode)
	log.Error(err.Error())
}
