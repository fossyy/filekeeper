package initialisation

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/middleware"
	"github.com/fossyy/filekeeper/types"
	"gorm.io/gorm"
	"io"
	"net/http"
	"os"
	"strconv"
)

var log *logger.AggregatedLogger

func init() {
	log = logger.Logger()
}

func POST(w http.ResponseWriter, r *http.Request) {

	session, _ := middleware.Store.Get(r, "session")
	userSession := middleware.GetUser(session)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		log.Error("Failed to read request body")
		return
	}

	var fileInfo types.FileInfo
	if err := json.Unmarshal(body, &fileInfo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}

	var currentInfo db.File
	err = db.DB.Table("files").Where("name = ? AND owner_id = ?", fileInfo.Name, userSession.UserID).First(&currentInfo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uploadDir := "uploads"
			if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
				err := os.Mkdir(uploadDir, os.ModePerm)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					log.Error(err.Error())
					return
				}
			}

			saveFolder := fmt.Sprintf("%s/%s/%s/tmp", uploadDir, userSession.UserID, fileInfo.Name)
			err = os.MkdirAll(saveFolder, os.ModePerm)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Error(err.Error())
				return
			}

			saveFolder = fmt.Sprintf("%s/%s/%s", uploadDir, userSession.UserID, fileInfo.Name)
			w.Header().Set("Content-Type", "application/json")

			if _, err := os.Stat(fmt.Sprintf("%s/info.json", saveFolder)); err == nil {
				open, _ := os.Open(fmt.Sprintf("%s/info.json", saveFolder))
				all, _ := io.ReadAll(open)
				var fileInfoUploaded types.FileInfoUploaded
				err := json.Unmarshal(all, &fileInfoUploaded)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					log.Error(err.Error())
					return
				}
				data := map[string]string{
					"status": strconv.Itoa(fileInfoUploaded.UploadedChunk),
				}
				err = json.NewEncoder(w).Encode(data)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					log.Error(err.Error())
					return
				}
				return
			} else if os.IsNotExist(err) {
				err := os.WriteFile(fmt.Sprintf("%s/info.json", saveFolder), body, 0644)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					log.Error(err.Error())
					return
				}
				data := map[string]string{
					"status": "ok",
				}
				err = json.NewEncoder(w).Encode(data)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					log.Error(err.Error())
					return
				}
				return
			}
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
	}
	data := map[string]string{
		"status": "conflict",
	}
	w.WriteHeader(http.StatusConflict)
	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
	return
}
