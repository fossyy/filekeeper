package renameFileHandler

import (
	"fmt"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	fileView "github.com/fossyy/filekeeper/view/client/file"
	"net/http"
	"strconv"
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

	prefix := fmt.Sprintf("%s/%s/chunk_", file.OwnerID.String(), file.ID.String())

	existingChunks, err := app.Server.Storage.ListObjects(r.Context(), prefix)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	missingChunk := len(existingChunks) != int(file.TotalChunk)

	fileData := types.FileData{
		ID:         newFile.ID.String(),
		Name:       newFile.Name,
		Size:       utils.ConvertFileSize(newFile.Size),
		IsPrivate:  newFile.IsPrivate,
		Type:       newFile.Type,
		Done:       !missingChunk,
		Downloaded: strconv.FormatUint(newFile.Downloaded, 10),
	}

	component := fileView.JustFile(fileData)
	err = component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	return
}
