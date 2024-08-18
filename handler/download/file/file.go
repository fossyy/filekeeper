package downloadFileHandler

import (
	"github.com/fossyy/filekeeper/app"
	"net/http"
	"os"
	"path/filepath"

	"github.com/fossyy/filekeeper/db"
)

func GET(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("id")
	file, err := db.DB.GetFile(fileID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	uploadDir := "uploads"

	currentDir, _ := os.Getwd()
	basePath := filepath.Join(currentDir, uploadDir)
	saveFolder := filepath.Join(basePath, file.OwnerID.String(), file.ID.String())

	if filepath.Dir(saveFolder) != filepath.Join(basePath, file.OwnerID.String()) {
		app.Server.Logger.Error("invalid path")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	openFile, err := os.OpenFile(filepath.Join(saveFolder, file.Name), os.O_RDONLY, 0)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	defer openFile.Close()

	stat, err := openFile.Stat()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+stat.Name())
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), openFile)
	return
}
