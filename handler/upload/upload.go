package uploadHandler

import (
	"encoding/json"
	"fmt"
	"github.com/fossyy/filekeeper/middleware"
	"github.com/fossyy/filekeeper/types"
	filesView "github.com/fossyy/filekeeper/view/upload"
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

	uploadDir := "uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, os.ModePerm)
	}
	fileName := r.FormValue("name")
	saveFolder := fmt.Sprintf("%s/%s/%s", uploadDir, userSession.UserID, fileName)

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
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		chunkFile.Close()

		if err := os.WriteFile(chunkName, fileData, 0644); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

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
	fmt.Println(r.FormValue("done"))
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
		fmt.Println(i)
	}
	outFile.Close()
	w.WriteHeader(http.StatusOK)
	return
}
