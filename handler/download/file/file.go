package downloadFileHandler

import (
	"fmt"
	"github.com/fossyy/filekeeper/db"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"path/filepath"
)

func GET(w http.ResponseWriter, r *http.Request) {
	fileID := mux.Vars(r)
	var file db.File
	db.DB.Table("files").Where("id = ?", fileID["id"]).First(&file)
	dir := fmt.Sprintf("uploads/%s/%s", file.OwnerID, file.Name)
	filePath := filepath.Join(dir, file.Name)
	openFile, _ := os.OpenFile(filePath, os.O_RDONLY, 0)
	stat, _ := openFile.Stat()
	w.Header().Set("Content-Disposition", "attachment; filename="+stat.Name())
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), openFile)
}
