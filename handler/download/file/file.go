package downloadFileHandler

import (
	"fmt"
	"github.com/fossyy/filekeeper/db"
	"github.com/gorilla/mux"
	"net/http"
	"os"
)

func GET(w http.ResponseWriter, r *http.Request) {
	fileID := mux.Vars(r)
	var file db.File
	db.DB.Table("files").Where("id = ?", fileID["id"]).First(&file)
	path := fmt.Sprintf("uploads/%s/%s/%s", file.OwnerID, file.Name, file.Name)
	openFile, _ := os.OpenFile(path, os.O_RDONLY, 0)
	stat, _ := openFile.Stat()
	http.ServeContent(w, r, openFile.Name(), stat.ModTime(), openFile)
}
