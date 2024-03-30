package uploadHandler

import (
	"encoding/json"
	"fmt"
	"github.com/fossyy/filekeeper/db"
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

func GET(w http.ResponseWriter, r *http.Request) {
	component := filesView.Main("upload page")
	component.Render(r.Context(), w)
}

func POST(w http.ResponseWriter, r *http.Request) {
	session, _ := middleware.Store.Get(r, "session")
	userSession := middleware.GetUser(session)

	r.ParseMultipartForm(10 << 20)

	fileName := r.FormValue("name")

	uploadDir := "uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, os.ModePerm)
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
	open.Close()
	var fileInfo types.FileInfo
	err = json.Unmarshal(all, &fileInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.FormValue("done") != "true" {
		chunkIndexStr := r.FormValue("index")
		chunkIndex, err := strconv.Atoi(chunkIndexStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		chunkFile, _, err := r.FormFile("chunk")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		chunkName := filepath.Join(fmt.Sprintf("%s/tmp", saveFolder), fmt.Sprintf("chunk_%d", chunkIndex))
		fileData, err := io.ReadAll(chunkFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		chunkFile.Close()

		if err := os.WriteFile(chunkName, fileData, 0644); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			return
		}

		os.WriteFile(fmt.Sprintf("%s/info.json", saveFolder), updatedJSON, 0644)

		w.WriteHeader(http.StatusOK)
		return
	}

	outFile, err := os.Create(fmt.Sprintf("%s/%s", saveFolder, fileInfo.Name))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		partFile.Close()

		err = os.Remove(fmt.Sprintf("%s/tmp/chunk_%d", saveFolder, i))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	outFile.Close()
	newFile := db.File{
		ID:         uuid.New(),
		OwnerID:    userSession.UserID,
		Name:       fileInfo.Name,
		Size:       fileInfo.Size,
		Downloaded: 0,
	}
	err = db.DB.Create(&newFile).Error
	if err != nil {
		fmt.Println(err.Error())
	}
	w.WriteHeader(http.StatusOK)
	return
}
