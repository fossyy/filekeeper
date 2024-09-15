package uploadHandler

import (
	"fmt"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/types"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func POST(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("id")
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userSession := r.Context().Value("user").(types.User)

	uploadDir := "uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.Mkdir(uploadDir, os.ModePerm); err != nil {
			app.Server.Logger.Error("error getting upload info: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	file, err := app.Server.Service.GetFile(fileID)
	if err != nil {
		app.Server.Logger.Error("error getting upload info: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rawIndex := r.FormValue("index")
	index, err := strconv.Atoi(rawIndex)
	if err != nil {
		return
	}

	currentDir, err := os.Getwd()
	if err != nil {
		app.Server.Logger.Error("unable to get current directory")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	basePath := filepath.Join(currentDir, uploadDir)
	cleanBasePath := filepath.Clean(basePath)

	saveFolder := filepath.Join(cleanBasePath, userSession.UserID.String(), file.ID.String(), file.Name)

	cleanSaveFolder := filepath.Clean(saveFolder)

	if !strings.HasPrefix(cleanSaveFolder, cleanBasePath) {
		app.Server.Logger.Error("invalid path")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := os.Stat(saveFolder); os.IsNotExist(err) {
		if err := os.MkdirAll(saveFolder, os.ModePerm); err != nil {
			app.Server.Logger.Error("error creating save folder: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	fileByte, _, err := r.FormFile("chunk")
	if err != nil {
		app.Server.Logger.Error("error getting upload info: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer fileByte.Close()

	dst, err := os.OpenFile(filepath.Join(saveFolder, fmt.Sprintf("chunk_%d", index)), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		app.Server.Logger.Error("error making upload folder: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer dst.Close()
	if _, err := io.Copy(dst, fileByte); err != nil {
		app.Server.Logger.Error("error copying byte to file dst: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	return
}
