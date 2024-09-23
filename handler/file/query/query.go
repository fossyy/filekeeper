package queryHandler

import (
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/types"
	fileView "github.com/fossyy/filekeeper/view/client/file"
	"net/http"
)

func GET(w http.ResponseWriter, r *http.Request) {
	userSession := r.Context().Value("user").(types.User)
	query := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")
	var fileStatus types.FileStatus

	if status == "private" {
		fileStatus = types.Private
	} else if status == "public" {
		fileStatus = types.Public
	} else {
		fileStatus = types.All
	}

	files, err := app.Server.Database.GetFiles(userSession.UserID.String(), query, fileStatus)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	var filesData []types.FileData

	for _, file := range files {
		userFile, err := app.Server.Service.GetUserFile(r.Context(), file.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			app.Server.Logger.Error(err.Error())
			return
		}

		filesData = append(filesData, *userFile)
	}

	if r.Header.Get("hx-request") == "true" {
		component := fileView.FileTable(filesData)
		err := component.Render(r.Context(), w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			app.Server.Logger.Error(err.Error())
			return
		}
		return
	}

	w.WriteHeader(http.StatusForbidden)
	return
}
