package deleteHandler

import (
	"fmt"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/types"
	"net/http"
)

func DELETE(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("id")
	consent := r.URL.Query().Get("consent")
	userSession := r.Context().Value("user").(types.User)

	file, err := app.Server.Database.GetFile(fileID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	if userSession.UserID != file.OwnerID || consent != "true" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = app.Server.Database.DeleteFile(fileID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	err = app.Server.Storage.Delete(r.Context(), fmt.Sprintf("%s/%s", file.OwnerID.String(), file.ID.String()))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}
