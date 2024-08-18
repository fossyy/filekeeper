package downloadHandler

import (
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/view/client/download"
	"net/http"

	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
)

func GET(w http.ResponseWriter, r *http.Request) {
	userSession := r.Context().Value("user").(types.User)
	files, err := db.DB.GetFiles(userSession.UserID.String())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var filesData []types.FileData
	for i := 0; i < len(files); i++ {
		filesData = append(filesData, types.FileData{
			ID:         files[i].ID.String(),
			Name:       files[i].Name,
			Size:       utils.ConvertFileSize(files[i].Size),
			Downloaded: files[i].Downloaded,
		})
	}

	component := downloadView.Main("Filekeeper - Download Page", filesData)
	err = component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
}
