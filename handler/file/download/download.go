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
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	status, userSession, _ := session.GetSession(r)
	if file.IsPrivate {
		if status == session.Unauthorized || status == session.InvalidSession || status == session.Suspicious {
			w.WriteHeader(http.StatusForbidden)
			return
		} else if file.OwnerID != userSession.UserID {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		start, end, valid := parseRange(rangeHeader, file.Size)
		if !valid {
			http.Error(w, "Invalid Range", http.StatusRequestedRangeNotSatisfiable)
			return
		}

		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, file.Size))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", end-start+1))
		w.WriteHeader(http.StatusPartialContent)
		sendFileChunk(r.Context(), w, file, start, end)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+file.Name)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Size))

	if file.Size > uint64(^uint64(0)>>1) {
		http.Error(w, "File is too large", http.StatusRequestedRangeNotSatisfiable)
		return
	}
	sendFileChunk(r.Context(), w, file, 0, int64(file.Size)-1)
	return
}

func parseRange(header string, size uint64) (start int64, end int64, valid bool) {
	if size == 0 || size > uint64(^uint64(0)>>1) {
		return 0, 0, false
	}
	rangeParts := strings.SplitN(strings.TrimPrefix(header, "bytes="), "-", 2)
	if len(rangeParts) != 2 {
		return 0, 0, false
	}

	start, err := strconv.ParseInt(rangeParts[0], 10, 64)
	if err != nil || start < 0 || start >= int64(size) {
		return 0, 0, false
	}

	end = int64(size) - 1
	if rangeParts[1] != "" {
		end, err = strconv.ParseInt(rangeParts[1], 10, 64)
		if err != nil || end < start {
			return 0, 0, false
		}
		if end >= int64(size) {
			end = int64(size) - 1
		}
	}

	return start, end, true
}

func sendFileChunk(ctx context.Context, w http.ResponseWriter, file *models.File, start, end int64) {
	chunkSize := int64(2 * 1024 * 1024)

	startChunk := start / chunkSize
	endChunk := end / chunkSize

	for i := startChunk; i <= endChunk; i++ {
		chunkKey := fmt.Sprintf("%s/%s/chunk_%d", file.OwnerID.String(), file.ID.String(), i)
		chunkData, err := app.Server.Storage.Get(ctx, chunkKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			app.Server.Logger.Error(err.Error())
			return
		}

		chunkLen := int64(len(chunkData))
		chunkStart := i * chunkSize
		startOffset := max(start-chunkStart, int64(0))
		endOffset := min(end-chunkStart, chunkSize-1)
		if chunkLen == 0 || startOffset >= chunkLen || endOffset < startOffset {
			continue
		}
		endOffset = min(endOffset, chunkLen-1)

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
			w.WriteHeader(http.StatusInternalServerError)
			app.Server.Logger.Error(err.Error())
			return
		}

		if i == int64(file.TotalChunk)-1 {
			err := app.Server.Database.IncrementDownloadCount(file.ID.String())
			if err != nil {
				app.Server.Logger.Error(err.Error())
				return
			}
		}
	}
}
