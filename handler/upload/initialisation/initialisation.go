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
			os.WriteFile(fmt.Sprintf("%s/info.json", saveFolder), body, 0644)
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusConflict)
}
