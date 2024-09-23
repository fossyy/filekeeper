package fileHandler

import (
	"fmt"
	"github.com/a-h/templ"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	fileView "github.com/fossyy/filekeeper/view/client/file"
	"net/http"
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
		userFile, err := app.Server.Service.GetUserFile(r.Context(), file.Name, file.OwnerID.String())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			app.Server.Logger.Error(err.Error())
			return
		}

		filesData = append(filesData, *userFile)
	}

	allowance, err := app.Server.Database.GetAllowance(userSession.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	usage, err := app.Server.Service.GetUserStorageUsage(r.Context(), userSession.UserID.String())
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
