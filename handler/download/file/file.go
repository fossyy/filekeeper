package downloadFileHandler

import (
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/types/models"
	"net/http"
	"os"
	"path/filepath"
)

var log *logger.AggregatedLogger

func init() {
	log = logger.Logger()
}

func GET(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("id")

	var file models.File
	err := db.DB.Table("files").Where("id = ?", fileID).First(&file).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
	}

	uploadDir := "uploads"

	currentDir, _ := os.Getwd()
	basePath := filepath.Join(currentDir, uploadDir)
	saveFolder := filepath.Join(basePath, file.OwnerID.String(), file.ID.String())

	if filepath.Dir(saveFolder) != filepath.Join(basePath, file.OwnerID.String()) {
		log.Error("invalid path")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	openFile, err := os.OpenFile(filepath.Join(saveFolder, file.Name), os.O_RDONLY, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
	}
	defer openFile.Close()

	stat, err := openFile.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+stat.Name())
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), openFile)
	return
}
