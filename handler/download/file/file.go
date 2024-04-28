package downloadFileHandler

import (
	"github.com/fossyy/filekeeper/utils"
	"net/http"
	"os"
	"path/filepath"

	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/logger"
)

var log *logger.AggregatedLogger

// TESTTING VAR
var database db.Database

func init() {
	log = logger.Logger()
	database = db.NewPostgresDB(utils.Getenv("DB_USERNAME"), utils.Getenv("DB_PASSWORD"), utils.Getenv("DB_HOST"), utils.Getenv("DB_PORT"), utils.Getenv("DB_NAME"))

}

func GET(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("id")
	file, err := database.GetFile(fileID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
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
		return
	}
	defer openFile.Close()

	stat, err := openFile.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+stat.Name())
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), openFile)
	return
}
