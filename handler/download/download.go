package downloadHandler

import (
	"errors"
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/middleware"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	downloadView "github.com/fossyy/filekeeper/view/download"
	"net/http"
)

var log *logger.AggregatedLogger

func init() {
	log = logger.Logger()
}

func GET(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("Session")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	storeSession, err := session.Store.Get(cookie.Value)
	if err != nil {
		if errors.Is(err, &session.SessionNotFound{}) {
			http.SetCookie(w, &http.Cookie{
				Name:   "Session",
				Value:  "",
				MaxAge: -1,
			})
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	userSession := middleware.GetUser(storeSession)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}

	var files []db.File
	db.DB.Table("files").Where("owner_id = ?", userSession.UserID).Find(&files)
	var filesData []types.FileData
	for i := 0; i < len(files); i++ {
		filesData = append(filesData, types.FileData{
			ID:         files[i].ID.String(),
			Name:       files[i].Name,
			Size:       utils.ConvertFileSize(files[i].Size),
			Downloaded: files[i].Downloaded,
		})
	}

	component := downloadView.Main("Download Page", filesData)
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
}
