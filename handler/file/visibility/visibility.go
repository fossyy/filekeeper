package visibilityHandler

import (
	"fmt"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	fileView "github.com/fossyy/filekeeper/view/client/file"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
	saveFolder := filepath.Join("uploads", userSession.UserID.String(), file.ID.String(), file.Name)
	missingChunk := false
	for j := 0; j < int(file.TotalChunk); j++ {
		fileName := fmt.Sprintf("%s/chunk_%d", saveFolder, j)

		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			missingChunk = true
			break
		}
	}
	fileData := types.FileData{
		ID:         file.ID.String(),
		Name:       file.Name,
		Size:       utils.ConvertFileSize(file.Size),
		IsPrivate:  !file.IsPrivate,
		Type:       file.Type,
		Done:       !missingChunk,
		Downloaded: strconv.FormatUint(file.Downloaded, 10),
	}
	component := fileView.JustFile(fileData)
	err = component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
