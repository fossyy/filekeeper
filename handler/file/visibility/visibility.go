package visibilityHandler

import (
	"fmt"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	fileView "github.com/fossyy/filekeeper/view/client/file"
	"net/http"
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
	saveFolder := filepath.Join("uploads", userSession.UserID.String(), file.ID.String())
	pattern := fmt.Sprintf("%s/chunk_*", saveFolder)
	chunkFiles, err := filepath.Glob(pattern)

	missingChunk := err != nil || len(chunkFiles) != int(file.TotalChunk)
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