package initialisation

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fossyy/filekeeper/app"
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

	fileData, err := app.Server.Service.GetUserFile(fileInfo.Name, userSession.UserID.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			upload, err := handleNewUpload(userSession, fileInfo)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			fileData = &types.FileWithDetail{
				ID:         fileData.ID,
				OwnerID:    fileData.OwnerID,
				Name:       fileData.Name,
				Size:       fileData.Size,
				Downloaded: fileData.Downloaded,
			}
			fileData.Chunk = make(map[string]bool)
			fileData.Done = true
			saveFolder := filepath.Join("uploads", userSession.UserID.String(), fileData.ID.String(), fileData.Name)
			for i := 0; i <= int(fileInfo.Chunk-1); i++ {
				fileName := fmt.Sprintf("%s/chunk_%d", saveFolder, i)

				if _, err := os.Stat(fileName); os.IsNotExist(err) {
					fileData.Chunk[fmt.Sprintf("chunk_%d", i)] = false
					fileData.Done = false
				} else {
					fileData.Chunk[fmt.Sprintf("chunk_%d", i)] = true
				}
			}
			respondJSON(w, upload)
			return
		}
		respondErrorJSON(w, err, http.StatusInternalServerError)
		return
	}

	fileData.Chunk = make(map[string]bool)
	fileData.Done = true
	saveFolder := filepath.Join("uploads", userSession.UserID.String(), fileData.ID.String(), fileData.Name)
	for i := 0; i <= int(fileInfo.Chunk-1); i++ {
		fileName := fmt.Sprintf("%s/chunk_%d", saveFolder, i)

		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			fileData.Chunk[fmt.Sprintf("chunk_%d", i)] = false
			fileData.Done = false
		} else {
			fileData.Chunk[fmt.Sprintf("chunk_%d", i)] = true
		}
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
		ID:         fileID,
		OwnerID:    ownerID,
		Name:       file.Name,
		Size:       file.Size,
		TotalChunk: file.Chunk - 1,
		Downloaded: 0,
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
