package uploadHandler

import (
	"encoding/json"
	"fmt"
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/middleware"
	"github.com/fossyy/filekeeper/types"
	filesView "github.com/fossyy/filekeeper/view/upload"
	"github.com/google/uuid"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

var log *logger.AggregatedLogger

func init() {
	log = logger.Logger()
}

func GET(w http.ResponseWriter, r *http.Request) {
	component := filesView.Main("upload page")
	err := component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
}

func POST(w http.ResponseWriter, r *http.Request) {
	session, _ := middleware.Store.Get(r, "session")
	userSession := middleware.GetUser(session)

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}

	fileName := r.FormValue("name")

	uploadDir := "uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		err := os.Mkdir(uploadDir, os.ModePerm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}
	}
	saveFolder := fmt.Sprintf("%s/%s/%s", uploadDir, userSession.UserID, fileName)

	open, err := os.Open(fmt.Sprintf("%s/info.json", saveFolder))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	all, err := io.ReadAll(open)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = open.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
	var fileInfo types.FileInfo
	err = json.Unmarshal(all, &fileInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}

	if r.FormValue("done") != "true" {
		chunkIndexStr := r.FormValue("index")
		chunkIndex, err := strconv.Atoi(chunkIndexStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}

		chunkFile, _, err := r.FormFile("chunk")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}

		chunkName := filepath.Join(fmt.Sprintf("%s/tmp", saveFolder), fmt.Sprintf("chunk_%d", chunkIndex))
		fileData, err := io.ReadAll(chunkFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}
		err = chunkFile.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}

		if err := os.WriteFile(chunkName, fileData, 0644); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}

		updatedFileInfo := types.FileInfoUploaded{
			Name:          fileInfo.Name,
			Size:          fileInfo.Size,
			Chunk:         fileInfo.Chunk,
			UploadedChunk: chunkIndex,
		}
		updatedJSON, err := json.Marshal(updatedFileInfo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}

		err = os.WriteFile(fmt.Sprintf("%s/info.json", saveFolder), updatedJSON, 0644)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	outFile, err := os.Create(fmt.Sprintf("%s/%s", saveFolder, fileInfo.Name))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}

	for i := 0; i <= fileInfo.Chunk-1; i += 1 {
		partFile, err := os.Open(fmt.Sprintf("%s/tmp/chunk_%d", saveFolder, i))
		if err != nil {
			panic(err)
		}
		_, err = io.Copy(outFile, partFile)
		if err != nil {
			panic(err)
		}
		err = partFile.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}

		err = os.Remove(fmt.Sprintf("%s/tmp/chunk_%d", saveFolder, i))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	err = outFile.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
	newFile := db.File{
		ID:         uuid.New(),
		OwnerID:    userSession.UserID,
		Name:       fileInfo.Name,
		Size:       fileInfo.Size,
		Downloaded: 0,
	}
	err = db.DB.Create(&newFile).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}
