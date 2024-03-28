package initialisation

import (
	"encoding/json"
	"fmt"
	"github.com/fossyy/filekeeper/middleware"
	"github.com/fossyy/filekeeper/types"
	"io"
	"net/http"
	"os"
)

func POST(w http.ResponseWriter, r *http.Request) {
	session, _ := middleware.Store.Get(r, "session")
	userSession := middleware.GetUser(session)
	fmt.Println("anjay")
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
}
