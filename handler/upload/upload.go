package uploadHandler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/fossyy/filekeeper/handler/upload/initialisation"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/middleware"
	"github.com/fossyy/filekeeper/session"
	filesView "github.com/fossyy/filekeeper/view/upload"
)

var log *logger.AggregatedLogger

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
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	cookie, err := r.Cookie("Session")
	if err != nil {
		handleCookieError(w, r, err)
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

	if r.FormValue("done") == "true" {
		fmt.Println("done")
		return
	}

	uploadID := r.FormValue("uploadID")

	uploadDir := "uploads"
	if err := createUploadDirectory(uploadDir); err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	file, err := initialisation.GetUploadInfo(uploadID)
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

	if err := updateIndex(r, uploadID); err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}
}

func createUploadDirectory(uploadDir string) error {
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.Mkdir(uploadDir, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func updateIndex(r *http.Request, uploadID string) error {
	rawIndex := r.FormValue("index")
	index, err := strconv.Atoi(rawIndex)
	if err != nil {
		return err
	}
	initialisation.UpdateIndex(uploadID, index)
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
