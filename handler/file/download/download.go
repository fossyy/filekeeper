package downloadHandler

import (
	"context"
	"fmt"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types/models"
	"net/http"
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

	status, userSession, _ := session.GetSession(r)
	if file.IsPrivate {
		if status == session.Unauthorized || status == session.InvalidSession {
			w.WriteHeader(http.StatusForbidden)
			return
		} else if file.OwnerID != userSession.UserID {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	//uploadDir := "uploads"
	//currentDir, _ := os.Getwd()
	//basePath := filepath.Join(currentDir, uploadDir)
	//saveFolder := filepath.Join(basePath, file.OwnerID.String(), file.ID.String())
	//
	//if filepath.Dir(saveFolder) != filepath.Join(basePath, file.OwnerID.String()) {
	//	http.Error(w, "Invalid Path", http.StatusInternalServerError)
	//	app.Server.Logger.Error("invalid path")
	//	return
	//}

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
			sendFileChunk(w, file, start, end)
			return
		}
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+file.Name)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Size))

	sendFileChunk(w, file, 0, int64(file.Size-1))
	return
}

func sendFileChunk(w http.ResponseWriter, file *models.File, start, end int64) {
	chunkSize := int64(2 * 1024 * 1024)

	startChunk := start / chunkSize
	endChunk := end / chunkSize

	startOffset := start % chunkSize
	endOffset := end % chunkSize

	for i := startChunk; i <= endChunk; i++ {
		chunkKey := fmt.Sprintf("%s/%s/chunk_%d", file.OwnerID.String(), file.ID.String(), i)
		chunkData, err := app.Server.Storage.Get(context.TODO(), chunkKey)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error retrieving chunk: %v", err), http.StatusInternalServerError)
			app.Server.Logger.Error(err.Error())
			return
		}

		var dataToSend []byte
		if i == startChunk && i == endChunk {
			dataToSend = chunkData[startOffset : endOffset+1]
		} else if i == startChunk {
			dataToSend = chunkData[startOffset:]
		} else if i == endChunk {
			dataToSend = chunkData[:endOffset+1]
		} else {
			dataToSend = chunkData
		}

		_, err = w.Write(dataToSend)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error writing chunk: %v", err), http.StatusInternalServerError)
			app.Server.Logger.Error(err.Error())
			return
		}

		if i == int64(file.TotalChunk)-1 {
			err := app.Server.Database.IncrementDownloadCount(file.ID.String())
			if err != nil {
				http.Error(w, fmt.Sprintf("Error updating download count: %v", err), http.StatusInternalServerError)
				app.Server.Logger.Error(err.Error())
				return
			}
		}
	}
}
