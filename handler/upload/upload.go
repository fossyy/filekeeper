package uploadHandler

import (
	"errors"
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/types"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/fossyy/filekeeper/logger"
	filesView "github.com/fossyy/filekeeper/view/upload"
)

var log *logger.AggregatedLogger
var mu sync.Mutex

func init() {
	log = logger.Logger()
}

func GET(w http.ResponseWriter, r *http.Request) {
	component := filesView.Main("upload page")
	if err := component.Render(r.Context(), w); err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}
}

func POST(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("id")
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	userSession := r.Context().Value("user").(types.User)

	if r.FormValue("done") == "true" {
		db.DB.FinalizeFileUpload(fileID)
		return
	}

	uploadDir := "uploads"
	if err := createUploadDirectory(uploadDir); err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	file, err := db.DB.GetUploadInfo(fileID)
	if err != nil {
		log.Error("error getting upload info: " + err.Error())
		return
	}

	currentDir, _ := os.Getwd()
	basePath := filepath.Join(currentDir, uploadDir)
	saveFolder := filepath.Join(basePath, userSession.UserID.String(), file.FileID.String())

	if filepath.Dir(saveFolder) != filepath.Join(basePath, userSession.UserID.String()) {
		log.Error("invalid path")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fileByte, _, err := r.FormFile("chunk")
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}
	defer fileByte.Close()

	dst, err := os.OpenFile(filepath.Join(saveFolder, file.Name), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, fileByte); err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}
	rawIndex := r.FormValue("index")
	index, err := strconv.Atoi(rawIndex)
	if err != nil {
		return
	}
	db.DB.UpdateUpdateIndex(index, fileID)
}

func createUploadDirectory(uploadDir string) error {
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.Mkdir(uploadDir, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func handleCookieError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, http.ErrNoCookie) {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}
	handleError(w, err, http.StatusInternalServerError)
}

func handleError(w http.ResponseWriter, err error, status int) {
	http.Error(w, err.Error(), status)
	log.Error(err.Error())
}
