package fileTableHandler

import (
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/types"
	fileView "github.com/fossyy/filekeeper/view/client/file"
	"net/http"
)

func GET(w http.ResponseWriter, r *http.Request) {
	userSession := r.Context().Value("user").(types.User)
	files, err := app.Server.Cache.GetUserFiles(r.Context(), userSession.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	var filesData []types.FileData

	for _, file := range files {
		userFile, err := app.Server.Cache.GetFileDetail(r.Context(), file.ID)
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
