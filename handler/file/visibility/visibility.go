package visibilityHandler

import (
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/types"
	fileView "github.com/fossyy/filekeeper/view/client/file"
	"net/http"
)

func PUT(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("id")
	userSession := r.Context().Value("user").(types.User)
	file, err := app.Server.Database.GetFile(fileID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	if file.OwnerID != userSession.UserID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err = app.Server.Database.ChangeFileVisibility(fileID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	err = app.Server.Service.DeleteFileCache(r.Context(), fileID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	userFile, err := app.Server.Service.GetUserFile(r.Context(), file.ID)
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
}
