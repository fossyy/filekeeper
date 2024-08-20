package initialisation

import (
	"encoding/json"
	"errors"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/cache"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/types/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func POST(w http.ResponseWriter, r *http.Request) {
	userSession := r.Context().Value("user").(types.User)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var fileInfo types.FileInfo
	if err := json.Unmarshal(body, &fileInfo); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fileData, err := cache.GetUserFile(fileInfo.Name, userSession.UserID.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			upload, err := handleNewUpload(userSession, fileInfo)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			respondJSON(w, upload)
			return
		}
		respondErrorJSON(w, err, http.StatusInternalServerError)
		return
	}

	if fileData.Done {
		respondJSON(w, map[string]bool{"Done": true})
		return
	}

	respondJSON(w, fileData)
}

func handleNewUpload(user types.User, file types.FileInfo) (models.File, error) {
	uploadDir := "uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		app.Server.Logger.Error(err.Error())
		err := os.Mkdir(uploadDir, os.ModePerm)
		if err != nil {
			app.Server.Logger.Error(err.Error())
			return models.File{}, err
		}
	}

	fileID := uuid.New()
	ownerID := user.UserID

	currentDir, _ := os.Getwd()
	basePath := filepath.Join(currentDir, uploadDir)
	saveFolder := filepath.Join(basePath, ownerID.String(), fileID.String())
	if filepath.Dir(saveFolder) != filepath.Join(basePath, ownerID.String()) {
		return models.File{}, errors.New("invalid path")
	}

	err := os.MkdirAll(saveFolder, os.ModePerm)
	if err != nil {
		app.Server.Logger.Error(err.Error())
		return models.File{}, err
	}

	newFile := models.File{
		ID:            fileID,
		OwnerID:       ownerID,
		Name:          file.Name,
		Size:          file.Size,
		Downloaded:    0,
		UploadedByte:  0,
		UploadedChunk: -1,
		Done:          false,
	}

	err = app.Server.Database.CreateFile(&newFile)
	if err != nil {
		app.Server.Logger.Error(err.Error())
		return models.File{}, err
	}

	return newFile, nil
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
	app.Server.Logger.Error(err.Error())
}
