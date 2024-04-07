package uploadHandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/middleware"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
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
	cookie, err := r.Cookie("Session")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	storeSession, err := session.Store.Get(cookie.Value)
	if err != nil {
		if errors.Is(err, &session.SessionNotFound{}) {
			storeSession.Destroy(w)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	userSession := middleware.GetUser(storeSession)

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}

	fileName := r.FormValue("name")
	fileName = utils.SanitizeFilename(fileName)

	uploadDir := "uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		err := os.Mkdir(uploadDir, os.ModePerm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}
	}

	currentDir, _ := os.Getwd()
	basePath := filepath.Join(currentDir, uploadDir)
	saveFolder := filepath.Join(basePath, userSession.UserID.String(), fileName)

	if filepath.Dir(saveFolder) != filepath.Join(basePath, userSession.UserID.String()) {
		log.Error("invalid path")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	open, err := os.Open(filepath.Join(saveFolder, "info.json"))
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

		err = os.WriteFile(filepath.Join(saveFolder, "info.json"), updatedJSON, 0644)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	outFile, err := os.Create(filepath.Join(saveFolder, fileInfo.Name))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}

	for i := 0; i <= fileInfo.Chunk-1; i += 1 {
		partFile, err := os.Open(filepath.Join(saveFolder, "tmp", fmt.Sprintf("chunk_%d", i)))
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

		err = os.Remove(filepath.Join(saveFolder, "tmp", fmt.Sprintf("chunk_%d", i)))
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
