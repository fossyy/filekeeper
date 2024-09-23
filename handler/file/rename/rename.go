package renameFileHandler

import (
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/types"
	fileView "github.com/fossyy/filekeeper/view/client/file"
	"net/http"
)

func PATCH(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("id")
	newName := r.URL.Query().Get("name")
	userSession := r.Context().Value("user").(types.User)

	file, err := app.Server.Database.GetFile(fileID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	if userSession.UserID != file.OwnerID {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if newName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newFile, err := app.Server.Database.RenameFile(fileID, newName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	userFile, err := app.Server.Service.GetUserFile(r.Context(), newFile.Name, newFile.OwnerID.String())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	component := fileView.JustFile(*userFile)
	err = component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	return
}
