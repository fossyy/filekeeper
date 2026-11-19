package deleteHandler

import (
	"errors"
	"fmt"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/cache"
	"github.com/fossyy/filekeeper/types"
	"gorm.io/gorm"
	"net/http"
)

func DELETE(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("id")
	consent := r.URL.Query().Get("consent")
	userSession := r.Context().Value("user").(types.User)

	file, err := app.Server.Database.GetFile(fileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	if userSession.UserID != file.OwnerID || consent != "true" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = app.Server.Storage.DeleteRecursive(r.Context(), fmt.Sprintf("%s/%s", file.OwnerID.String(), file.ID.String()))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	err = app.Server.Database.DeleteFile(fileID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	err = app.Server.Cache.RemoveUserFilesCache(r.Context(), userSession.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	err = app.Server.Cache.RemoveFileCache(r.Context(), fileID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	err = app.Server.Cache.DeleteCache(r.Context(), fmt.Sprintf(cache.FileChunkCacheKey, fileID))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}
