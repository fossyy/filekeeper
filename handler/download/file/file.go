package downloadFileHandler

import (
	"fmt"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/types/models"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func GET(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("id")
	file, err := app.Server.Database.GetFile(fileID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	uploadDir := "uploads"
	currentDir, _ := os.Getwd()
	basePath := filepath.Join(currentDir, uploadDir)
	saveFolder := filepath.Join(basePath, file.OwnerID.String(), file.ID.String())

	if filepath.Dir(saveFolder) != filepath.Join(basePath, file.OwnerID.String()) {
		http.Error(w, "Invalid Path", http.StatusInternalServerError)
		app.Server.Logger.Error("invalid path")
		return
	}

	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		rangeParts := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")
		if len(rangeParts) == 2 {
			start, err := strconv.ParseInt(rangeParts[0], 10, 64)
			if err != nil {
				http.Error(w, "Invalid Range", http.StatusRequestedRangeNotSatisfiable)
				return
			}
			end := int64(file.Size - 1)
			if rangeParts[1] != "" {
				end, err = strconv.ParseInt(rangeParts[1], 10, 64)
				if err != nil {
					http.Error(w, "Invalid Range", http.StatusRequestedRangeNotSatisfiable)
					return
				}
			}

			if end >= int64(file.Size) {
				end = int64(file.Size - 1)
			}

			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, file.Size))
			w.Header().Set("Content-Length", fmt.Sprintf("%d", end-start+1))
			w.WriteHeader(http.StatusPartialContent)
			sendFileChunk(w, saveFolder, file, start, end)
			return
		}
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+file.Name)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Size))

	sendFileChunk(w, saveFolder, file, 0, int64(file.Size-1))
}

func sendFileChunk(w http.ResponseWriter, saveFolder string, file *models.File, start, end int64) {
	chunkSize := int64(2 * 1024 * 1024)

	startChunk := start / chunkSize
	endChunk := end / chunkSize

	startOffset := start % chunkSize
	endOffset := end % chunkSize

	for i := startChunk; i <= endChunk; i++ {
		chunkPath := filepath.Join(saveFolder, file.Name, fmt.Sprintf("chunk_%d", i))
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error opening chunk: %v", err), http.StatusInternalServerError)
			app.Server.Logger.Error(err.Error())
			return
		}
		defer chunkFile.Close()

		var chunkStart, chunkEnd int64
		if i == startChunk {
			chunkStart = startOffset
		} else {
			chunkStart = 0
		}
		if i == endChunk {
			chunkEnd = endOffset
		} else {
			chunkEnd = chunkSize - 1
		}

		_, err = chunkFile.Seek(chunkStart, io.SeekStart)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error seeking chunk: %v", err), http.StatusInternalServerError)
			app.Server.Logger.Error(err.Error())
			return
		}

		buffer := make([]byte, 2048)
		toSend := chunkEnd - chunkStart + 1
		for toSend > 0 {
			n, err := chunkFile.Read(buffer)
			if err != nil && err != io.EOF {
				http.Error(w, fmt.Sprintf("Error reading chunk: %v", err), http.StatusInternalServerError)
				app.Server.Logger.Error(err.Error())
				return
			}
			if n == 0 {
				break
			}
			if int64(n) > toSend {
				n = int(toSend)
			}
			_, err = w.Write(buffer[:n])
			if err != nil {
				http.Error(w, fmt.Sprintf("Error writing chunk: %v", err), http.StatusInternalServerError)
				app.Server.Logger.Error(err.Error())
				return
			}
			toSend -= int64(n)
		}
	}
}
