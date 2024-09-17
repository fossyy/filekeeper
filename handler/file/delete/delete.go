package deleteHandler

import (
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/types"
	"net/http"
	"os"
	"path/filepath"
)

func DELETE(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("id")
	consent := r.URL.Query().Get("consent")
	userSession := r.Context().Value("user").(types.User)

	file, err := app.Server.Database.GetFile(fileID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if userSession.UserID != file.OwnerID || consent != "true" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = app.Server.Database.DeleteFile(fileID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	uploadDir := "uploads"
	currentDir, _ := os.Getwd()
	basePath := filepath.Join(currentDir, uploadDir)
	fileFolder := filepath.Join(basePath, file.OwnerID.String(), file.ID.String())
	err = os.RemoveAll(fileFolder)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}
