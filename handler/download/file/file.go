package downloadFileHandler

import (
	"fmt"
	"github.com/fossyy/filekeeper/app"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func GET(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("id")
	file, err := app.Server.Database.GetFile(fileID)
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

	w.Header().Set("Content-Disposition", "attachment; filename="+file.Name)
	w.Header().Set("Content-Type", "application/octet-stream")
	for i := 0; i <= int(file.TotalChunk); i++ {
		chunkPath := filepath.Join(saveFolder, file.Name, fmt.Sprintf("chunk_%d", i))

		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error opening chunk: %v", err), http.StatusInternalServerError)
			return
		}
		_, err = io.Copy(w, chunkFile)
		chunkFile.Close()
		if err != nil {
			http.Error(w, fmt.Sprintf("Error writing chunk: %v", err), http.StatusInternalServerError)
			return
		}
	}
}
