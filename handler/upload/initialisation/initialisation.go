package initialisation

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/middleware"
	"github.com/fossyy/filekeeper/types"
	"gorm.io/gorm"
	"io"
	"net/http"
	"os"
	"strconv"
)

func POST(w http.ResponseWriter, r *http.Request) {
	session, _ := middleware.Store.Get(r, "session")
	userSession := middleware.GetUser(session)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	var fileInfo types.FileInfo
	if err := json.Unmarshal(body, &fileInfo); err != nil {
		fmt.Println(err.Error())
		return
	}

	var currentInfo db.File
	err = db.DB.Table("files").Where("name = ? AND owner_id = ?", fileInfo.Name, userSession.UserID).First(&currentInfo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uploadDir := "uploads"
			if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
				os.Mkdir(uploadDir, os.ModePerm)
			}

			saveFolder := fmt.Sprintf("%s/%s/%s/tmp", uploadDir, userSession.UserID, fileInfo.Name)
			err = os.MkdirAll(saveFolder, os.ModePerm)
			if err != nil {
				fmt.Println(err.Error())
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
					fmt.Println(err.Error())
					return
				}
				data := map[string]string{
					"status": strconv.Itoa(fileInfoUploaded.UploadedChunk),
				}
				json.NewEncoder(w).Encode(data)
				return
			} else if os.IsNotExist(err) {
				os.WriteFile(fmt.Sprintf("%s/info.json", saveFolder), body, 0644)
				data := map[string]string{
					"status": "ok",
				}
				json.NewEncoder(w).Encode(data)
				return
			}
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	data := map[string]string{
		"status": "conflict",
	}
	w.WriteHeader(http.StatusConflict)
	json.NewEncoder(w).Encode(data)
}
