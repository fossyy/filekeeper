package fileHandler

import (
	"fmt"
	"github.com/a-h/templ"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	fileView "github.com/fossyy/filekeeper/view/client/file"
	"net/http"
	"strconv"
)

func GET(w http.ResponseWriter, r *http.Request) {
	userSession := r.Context().Value("user").(types.User)
	files, err := app.Server.Database.GetFiles(userSession.UserID.String(), "", types.All)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	var filesData []types.FileData

	for _, file := range files {
		prefix := fmt.Sprintf("%s/%s/chunk_", file.OwnerID.String(), file.ID.String())

		existingChunks, err := app.Server.Storage.ListObjects(r.Context(), prefix)
		if err != nil {
			app.Server.Logger.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		missingChunk := len(existingChunks) != int(file.TotalChunk)

		filesData = append(filesData, types.FileData{
			ID:         file.ID.String(),
			Name:       file.Name,
			Size:       utils.ConvertFileSize(file.Size),
			IsPrivate:  file.IsPrivate,
			Type:       file.Type,
			Done:       !missingChunk,
			Downloaded: strconv.FormatUint(file.Downloaded, 10),
		})
	}

	allowance, err := app.Server.Database.GetAllowance(userSession.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	usage, err := app.Server.Service.GetUserStorageUsage(userSession.UserID.String())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	allowanceStats := &types.Allowance{
		AllowanceByte:        utils.ConvertFileSize(allowance.AllowanceByte),
		AllowanceUsedByte:    utils.ConvertFileSize(usage),
		AllowanceUsedPercent: fmt.Sprintf("%.2f", float64(usage)/float64(allowance.AllowanceByte)*100),
	}

	var component templ.Component
	if r.Header.Get("hx-request") == "true" {
		component = fileView.MainContent("Filekeeper - File Dashboard", filesData, userSession, allowanceStats)
	} else {
		component = fileView.Main("Filekeeper - File Dashboard", filesData, userSession, allowanceStats)
	}

	err = component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
}
