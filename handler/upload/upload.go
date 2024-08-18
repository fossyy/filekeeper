package uploadHandler

import (
	"github.com/fossyy/filekeeper/cache"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/types"
	filesView "github.com/fossyy/filekeeper/view/client/upload"
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
	component := filesView.Main("Filekeeper - Upload")
	if err := component.Render(r.Context(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

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
			log.Error("error getting upload info: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	file, err := cache.GetFile(fileID)
	if err != nil {
		log.Error("error getting upload info: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	currentDir, _ := os.Getwd()
	basePath := filepath.Join(currentDir, uploadDir)
	saveFolder := filepath.Join(basePath, userSession.UserID.String(), file.ID.String())

	if filepath.Dir(saveFolder) != filepath.Join(basePath, userSession.UserID.String()) {
		log.Error("invalid path")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fileByte, fileHeader, err := r.FormFile("chunk")
	if err != nil {
		log.Error("error getting upload info: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer fileByte.Close()

	rawIndex := r.FormValue("index")
	index, err := strconv.Atoi(rawIndex)
	if err != nil {
		return
	}

	file.UpdateProgress(int64(index), file.UploadedByte+int64(fileHeader.Size))

	dst, err := os.OpenFile(filepath.Join(saveFolder, file.Name), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Error("error making upload folder: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer dst.Close()
	if _, err := io.Copy(dst, fileByte); err != nil {
		log.Error("error copying byte to file dst: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if file.UploadedByte >= file.Size {
		file.FinalizeFileUpload()
		return
	}
	return
}
