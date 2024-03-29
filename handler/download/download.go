package downloadHandler

import (
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/middleware"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	downloadView "github.com/fossyy/filekeeper/view/download"
	"net/http"
)

func GET(w http.ResponseWriter, r *http.Request) {
	session, err := middleware.Store.Get(r, "session")
	userSession := middleware.GetUser(session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	component.Render(r.Context(), w)
}
